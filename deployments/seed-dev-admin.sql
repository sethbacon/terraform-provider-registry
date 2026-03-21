-- Seed a dev admin user with org membership.
-- Run after the backend has started (migrations must have run first).
DO $$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
    v_admin_role_template_id uuid;
BEGIN
    INSERT INTO users (email, name, oidc_sub)
    VALUES ('admin@dev.local', 'Dev Admin', 'dev-admin-oidc-sub')
    ON CONFLICT (email) DO NOTHING;

    SELECT id INTO v_user_id FROM users WHERE email = 'admin@dev.local';
    SELECT id INTO v_org_id FROM organizations WHERE name = 'default';
    SELECT id INTO v_admin_role_template_id FROM role_templates WHERE name = 'admin';

    INSERT INTO organization_members (organization_id, user_id, role_template_id)
    VALUES (v_org_id, v_user_id, v_admin_role_template_id)
    ON CONFLICT (organization_id, user_id)
    DO UPDATE SET role_template_id = EXCLUDED.role_template_id;
END $$;
