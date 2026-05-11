#!/usr/bin/env python3
"""
preprocess.py — Patch the backend's openapi3.json so strict OpenAPI 3 validators
(notably oapi-codegen) accept it.

The backend emits path parameters at the operation level only, which is
OpenAPI 3 compliant but rejected by oapi-codegen for templated paths. This
script lifts operation-level path params to the path level for every
templated path, leaving the underlying spec semantics unchanged.

Once terraform-registry-backend#359 lands, this preprocessor can be deleted.

Usage:
    python3 preprocess.py <input.json> <output.json>
"""
import json
import re
import sys


def declare_missing_security_schemes(spec):
    """Backend operations reference a SetupToken security scheme but the
    components.securitySchemes block only declares Bearer. oapi-codegen emits
    a `SetupTokenScopes` const referencing an undeclared type. Add the missing
    scheme so the generated code compiles.

    SetupToken is a Bearer-style token (header: X-Setup-Token). It's only
    used in /api/v1/setup/* paths and only at install time, but it has to
    be declared for the spec to validate strictly.
    """
    components = spec.setdefault('components', {})
    schemes = components.setdefault('securitySchemes', {})
    if 'SetupToken' not in schemes:
        schemes['SetupToken'] = {
            'type': 'apiKey',
            'name': 'X-Setup-Token',
            'in': 'header',
            'description': 'One-time setup token, valid until /api/v1/setup/complete is called.',
        }
    return spec


def dedupe_enum_values(spec):
    """Deduplicate enum value lists where the same string value appears twice.

    The backend declares some Go enum aliases (e.g. `Kind* = Provider*` in
    scm/types.go) which swag emits as duplicate enum values. oapi-codegen
    then generates duplicate-case switches that won't compile. Tracked as
    terraform-registry-backend#360.
    """
    schemas = spec.get('components', {}).get('schemas', {})
    for schema in schemas.values():
        if not isinstance(schema, dict):
            continue
        if schema.get('type') != 'string':
            continue
        values = schema.get('enum')
        if not values:
            continue
        seen = set()
        deduped_values = []
        deduped_names = []
        names = schema.get('x-enum-varnames') or []
        for i, v in enumerate(values):
            if v in seen:
                continue
            seen.add(v)
            deduped_values.append(v)
            if i < len(names):
                deduped_names.append(names[i])
        schema['enum'] = deduped_values
        if 'x-enum-varnames' in schema:
            schema['x-enum-varnames'] = deduped_names
    return spec


def lift_path_params(spec):
    """Promote operation-level path parameters to path level."""
    methods = ('get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace')
    paths = spec.get('paths', {})

    for path, path_item in paths.items():
        if '{' not in path:
            continue
        if 'parameters' in path_item:
            # Path-level params already present; trust them.
            continue

        placeholders = set(re.findall(r'\{([^}]+)\}', path))
        # Collect the first matching path-param spec from each operation,
        # keyed by name. If multiple operations declare different shapes for
        # the same param, the last one wins — that's safer than failing,
        # since the backend should keep them in sync.
        params_by_name = {}
        for method in methods:
            op = path_item.get(method)
            if not isinstance(op, dict):
                continue
            for p in op.get('parameters', []):
                if p.get('in') == 'path' and p.get('name') in placeholders:
                    params_by_name[p['name']] = p

        if not params_by_name:
            # No operation declared any path params. Synthesize string-typed
            # ones from the placeholders so the spec at least validates.
            params_by_name = {
                name: {
                    'name': name,
                    'in': 'path',
                    'required': True,
                    'schema': {'type': 'string'},
                }
                for name in placeholders
            }

        # Cover any placeholders the operations missed.
        for missing in placeholders - set(params_by_name.keys()):
            params_by_name[missing] = {
                'name': missing,
                'in': 'path',
                'required': True,
                'schema': {'type': 'string'},
            }

        path_item['parameters'] = [params_by_name[n] for n in sorted(params_by_name.keys())]

    return spec


def main():
    if len(sys.argv) != 3:
        print('usage: preprocess.py <input.json> <output.json>', file=sys.stderr)
        sys.exit(2)
    src, dst = sys.argv[1], sys.argv[2]
    with open(src, encoding='utf-8') as f:
        spec = json.load(f)
    spec = declare_missing_security_schemes(spec)
    spec = dedupe_enum_values(spec)
    spec = lift_path_params(spec)
    with open(dst, 'w', encoding='utf-8') as f:
        json.dump(spec, f, indent=2)
    paths = spec.get('paths', {})
    lifted = sum(1 for p, item in paths.items() if '{' in p and 'parameters' in item)
    schemas = spec.get('components', {}).get('schemas', {})
    print(f'preprocessed: {len(paths)} paths total, {lifted} have path-level parameters, '
          f'{len(schemas)} schemas (deduped string enums in place)')


if __name__ == '__main__':
    main()
