package sdk

import (
	"log"

	"github.com/clintjedwards/tfvet/internal/cli/models"
	tfvetPlugin "github.com/clintjedwards/tfvet/internal/plugin"
	proto "github.com/clintjedwards/tfvet/internal/plugin/proto"
	"github.com/hashicorp/go-plugin"
)

// Check provides an interface for the user to define their own check/lint method.
// This is the core of the pluggable interface pattern and allows the user to simply consume
// the hcl file and return linting errors.
//
// content is the full hclfile in byte format.
type Check interface {
	Check(content []byte) ([]RuleError, error)
}

// Rule is the representation of a single rule within tfvet.
// This just combines the rule
type Rule struct {
	models.Rule
	Check
}

// Position represents location within a document.
type Position struct {
	Line   uint32
	Column uint32
}

// Range represents the starting and ending points on a specific line within a document.
type Range struct {
	Start Position
	End   Position
}

// RuleError represents a single lint error's details
type RuleError struct {
	Suggestion  string
	Remediation string
	Location    Range
	// metadata is a key value store that allows the rule to include extra data,
	// that can be used by any tooling consuming said rule. For example "severity"
	// might be something included in metadata.
	Metadata map[string]string
}

// GetRuleInfo returns information about the rule itself.
func (rule *Rule) GetRuleInfo(request *proto.GetRuleInfoRequest) (*proto.GetRuleInfoResponse, error) {
	ruleInfo := proto.GetRuleInfoResponse{
		RuleInfo: &proto.RuleInfo{
			Name:    rule.Name,
			Short:   rule.Short,
			Long:    rule.Long,
			Link:    rule.Link,
			Enabled: rule.Enabled,
		},
	}

	return &ruleInfo, nil
}

// ExecuteRule runs the linting rule given a single file and returns any linting errors.
func (rule *Rule) ExecuteRule(request *proto.ExecuteRuleRequest) (*proto.ExecuteRuleResponse, error) {
	ruleErrors, err := rule.Check.Check(request.HclFile)

	return &proto.ExecuteRuleResponse{
		Errors: ruleErrorsToProto(ruleErrors),
	}, err
}

func ruleErrorsToProto(ruleErrors []RuleError) []*proto.RuleError {
	protoRuleErrors := []*proto.RuleError{}

	for _, ruleError := range ruleErrors {
		protoRuleErrors = append(protoRuleErrors, &proto.RuleError{
			Location: &proto.Location{
				Start: &proto.Position{
					Line:   ruleError.Location.Start.Line,
					Column: ruleError.Location.Start.Column,
				},
				End: &proto.Position{
					Line:   ruleError.Location.End.Line,
					Column: ruleError.Location.End.Column,
				},
			},
			Suggestion:  ruleError.Suggestion,
			Remediation: ruleError.Remediation,
			Metadata:    ruleError.Metadata,
		})
	}

	return protoRuleErrors
}

// validates a new rule has at least the basic information
func (rule *Rule) isValid() bool {
	if rule.Short == "" {
		return false
	}

	if rule.Name == "" {
		return false
	}

	if rule.Check == nil {
		return false
	}

	return true
}

// NewRule registers a new linting rule. This function must be included inside a rule.
func NewRule(rule *Rule) {
	if !rule.isValid() {
		log.Fatalf("%s is not valid", rule.Name)
		return
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tfvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			// The key here is to enable different plugins to be served by one binary
			"tfvet-sdk": &tfvetPlugin.TfvetRulePlugin{Impl: rule},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
