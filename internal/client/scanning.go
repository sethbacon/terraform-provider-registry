package client

import (
	"context"
	"fmt"
)

// GetScanningConfig fetches the scanner configuration via GET /api/v1/admin/scanning/config.
func (c *Client) GetScanningConfig(ctx context.Context) (*ScanningConfig, error) {
	var cfg ScanningConfig
	if err := c.Get(ctx, "/api/v1/admin/scanning/config", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetScanningStats fetches aggregate scan statistics via GET /api/v1/admin/scanning/stats.
func (c *Client) GetScanningStats(ctx context.Context) (*ScanningStats, error) {
	var stats ScanningStats
	if err := c.Get(ctx, "/api/v1/admin/scanning/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetScan fetches a scan by UUID via GET /api/v1/admin/scanning/scans/{id}.
func (c *Client) GetScan(ctx context.Context, id string) (*Scan, error) {
	var s Scan
	if err := c.Get(ctx, "/api/v1/admin/scanning/scans/"+id, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetModuleScan fetches the most recent scan for a module version.
func (c *Client) GetModuleScan(ctx context.Context, namespace, name, system, version string) (*Scan, error) {
	var s Scan
	path := fmt.Sprintf("/api/v1/modules/%s/%s/%s/versions/%s/scan", namespace, name, system, version)
	if err := c.Get(ctx, path, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
