package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
)

// ScanFindingModel is the Terraform model for a scan finding.
type ScanFindingModel struct {
	ID          types.String `tfsdk:"id"`
	RuleID      types.String `tfsdk:"rule_id"`
	Severity    types.String `tfsdk:"severity"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Resource    types.String `tfsdk:"resource"`
	FilePath    types.String `tfsdk:"file_path"`
	LineNumber  types.Int64  `tfsdk:"line_number"`
}

// ScanDataModel holds the common attributes of a scan result.
type ScanDataModel struct {
	ID           types.String       `tfsdk:"id"`
	Status       types.String       `tfsdk:"status"`
	Scanner      types.String       `tfsdk:"scanner"`
	Passed       types.Bool         `tfsdk:"passed"`
	Findings     []ScanFindingModel `tfsdk:"findings"`
	ExecutionLog types.String       `tfsdk:"execution_log"`
	StartedAt    types.String       `tfsdk:"started_at"`
	CompletedAt  types.String       `tfsdk:"completed_at"`
	CreatedAt    types.String       `tfsdk:"created_at"`
}

func scanToModel(s *client.Scan) ScanDataModel {
	m := ScanDataModel{
		ID:        types.StringValue(s.ID),
		Status:    types.StringValue(s.Status),
		Scanner:   types.StringValue(s.Scanner),
		Passed:    types.BoolValue(s.Passed),
		CreatedAt: types.StringValue(normalizeTimestamp(s.CreatedAt)),
	}
	if s.ExecutionLog != nil {
		m.ExecutionLog = types.StringValue(*s.ExecutionLog)
	} else {
		m.ExecutionLog = types.StringValue("")
	}
	if s.StartedAt != nil {
		m.StartedAt = types.StringValue(normalizeTimestamp(*s.StartedAt))
	} else {
		m.StartedAt = types.StringValue("")
	}
	if s.CompletedAt != nil {
		m.CompletedAt = types.StringValue(normalizeTimestamp(*s.CompletedAt))
	} else {
		m.CompletedAt = types.StringValue("")
	}

	findings := make([]ScanFindingModel, 0, len(s.Findings))
	for _, f := range s.Findings {
		fm := ScanFindingModel{
			ID:       types.StringValue(f.ID),
			RuleID:   types.StringValue(f.RuleID),
			Severity: types.StringValue(f.Severity),
			Title:    types.StringValue(f.Title),
		}
		opt := func(s *string) types.String {
			if s != nil {
				return types.StringValue(*s)
			}
			return types.StringValue("")
		}
		fm.Description = opt(f.Description)
		fm.Resource = opt(f.Resource)
		fm.FilePath = opt(f.FilePath)
		if f.LineNumber != nil {
			fm.LineNumber = types.Int64Value(int64(*f.LineNumber))
		} else {
			fm.LineNumber = types.Int64Value(0)
		}
		findings = append(findings, fm)
	}
	m.Findings = findings
	return m
}
