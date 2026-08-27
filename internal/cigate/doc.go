// Package cigate holds assertions about this repository's own CI wiring.
//
// Nothing here ships in the provider. It exists because some of what keeps this
// repository safe is a relationship between two workflow jobs rather than a
// property of any one file, and a relationship joined only by a comment is one
// nothing checks. See osv_reachability_test.go.
package cigate
