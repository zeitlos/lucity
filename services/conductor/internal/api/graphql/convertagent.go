package graphql

import (
	"strings"

	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/agentaudit"
	"github.com/zeitlos/lucity/services/conductor/internal/agentrunner"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

func parseModelAPI(api model.ModelAPI) (values.ModelAPI, error) {
	switch api {
	case model.ModelAPIAnthropic:
		return values.ModelAPIAnthropic, nil
	case model.ModelAPIOpenai:
		return values.ModelAPIOpenAI, nil
	case model.ModelAPIOpenaiCompatible:
		return values.ModelAPIOpenAICompatible, nil
	default:
		return "", usererror.Errorf("unsupported api %q", api)
	}
}

func convertModelAPI(api values.ModelAPI) model.ModelAPI {
	switch api {
	case values.ModelAPIAnthropic:
		return model.ModelAPIAnthropic
	case values.ModelAPIOpenAI:
		return model.ModelAPIOpenai
	default:
		return model.ModelAPIOpenaiCompatible
	}
}

func convertCredentialStatus(status values.CredentialStatus) model.CredentialStatus {
	switch status {
	case values.CredentialValid:
		return model.CredentialStatusValid
	case values.CredentialRejected:
		return model.CredentialStatusRejected
	default:
		return model.CredentialStatusUnreachable
	}
}

func convertModelProvider(provider conductor.ModelProvider) model.ModelProvider {
	models := provider.Models

	if models == nil {
		models = []string{}
	}

	return model.ModelProvider{
		ID:        provider.ID,
		Name:      provider.ID.Name,
		API:       convertModelAPI(provider.API),
		URL:       provider.URL,
		KeyHint:   optional(provider.KeyHint),
		Status:    convertCredentialStatus(provider.Status),
		CheckedAt: provider.CheckedAt,
		Models:    models,
	}
}

func convertConnector(connector conductor.Connector) model.Connector {
	tools := make([]model.ConnectorTool, 0, len(connector.Tools))

	for _, tool := range connector.Tools {
		tools = append(tools, model.ConnectorTool{Name: tool.Name, Description: optional(tool.Description)})
	}

	return model.Connector{
		ID:             connector.ID,
		Name:           connector.ID.Name,
		URL:            connector.URL,
		CredentialHint: optional(connector.CredentialHint),
		Status:         convertCredentialStatus(connector.Status),
		CheckedAt:      connector.CheckedAt,
		Tools:          tools,
	}
}

func parseHarness(harness model.Harness) (values.Harness, error) {
	switch harness {
	case model.HarnessClaudeCode:
		return values.HarnessClaudeCode, nil
	case model.HarnessCodex:
		return values.HarnessCodex, nil
	case model.HarnessOpencode:
		return values.HarnessOpenCode, nil
	default:
		return "", usererror.Errorf("unsupported harness %q", harness)
	}
}

func convertHarness(harness values.Harness) model.Harness {
	switch harness {
	case values.HarnessCodex:
		return model.HarnessCodex
	case values.HarnessOpenCode:
		return model.HarnessOpencode
	default:
		return model.HarnessClaudeCode
	}
}

func parseApproval(approval model.Approval) (values.Approval, error) {
	switch approval {
	case model.ApprovalAllow:
		return values.ApprovalAllow, nil
	case model.ApprovalAsk:
		return values.ApprovalAsk, nil
	default:
		return "", usererror.Errorf("unsupported approval %q", approval)
	}
}

func convertApproval(approval values.Approval) model.Approval {
	if approval == values.ApprovalAllow {
		return model.ApprovalAllow
	}

	return model.ApprovalAsk
}

func parseAgentConnectors(inputs []model.AgentConnectorInput) ([]conductor.AgentConnectorSpec, error) {
	specs := make([]conductor.AgentConnectorSpec, 0, len(inputs))

	for _, input := range inputs {
		approval, err := parseApproval(input.Approval)

		if err != nil {
			return nil, err
		}

		specs = append(specs, conductor.AgentConnectorSpec{
			Connector:   input.Connector,
			Approval:    approval,
			HiddenTools: input.HiddenTools,
		})
	}

	return specs, nil
}

func convertAgent(agent conductor.Agent) model.Agent {
	connectors := make([]model.AgentConnector, 0, len(agent.Connectors))

	for _, granted := range agent.Connectors {
		hidden := granted.HiddenTools

		if hidden == nil {
			hidden = []string{}
		}

		connectors = append(connectors, model.AgentConnector{
			Connector:   new(convertConnector(granted.Connector)),
			Approval:    convertApproval(granted.Approval),
			HiddenTools: hidden,
		})
	}

	allowedHosts := agent.AllowedHosts

	if allowedHosts == nil {
		allowedHosts = []string{}
	}

	return model.Agent{
		ID:             agent.ID,
		Name:           agent.ID.Name,
		Harness:        convertHarness(agent.Harness),
		Instructions:   agent.Instructions,
		Connectors:     connectors,
		AllowedHosts:   allowedHosts,
		PrivateNetwork: agent.PrivateNetwork,
	}
}

func convertAgents(agents []conductor.Agent) []model.Agent {
	result := make([]model.Agent, 0, len(agents))

	for _, agent := range agents {
		result = append(result, convertAgent(agent))
	}

	return result
}

func convertAgentStatus(status conductor.AgentStatus) model.AgentStatus {
	switch status {
	case conductor.AgentIdle:
		return model.AgentStatusIdle
	case conductor.AgentWorking:
		return model.AgentStatusWorking
	case conductor.AgentWaitingForApproval:
		return model.AgentStatusWaitingForApproval
	case conductor.AgentFailed:
		return model.AgentStatusFailed
	default:
		return model.AgentStatusStarting
	}
}

func convertAgentModel(agentModel conductor.AgentModel) model.AgentModel {
	return model.AgentModel{
		Provider: new(convertModelProvider(agentModel.Provider)),
		Name:     agentModel.Name,
	}
}

func parseAgentModel(input model.AgentModelInput) conductor.AgentModelSpec {
	return conductor.AgentModelSpec{Provider: input.Provider, Name: input.Name}
}

func parseEffort(effort model.Effort) string {
	return strings.ToLower(string(effort))
}

func convertEffort(effort string) model.Effort {
	converted := model.Effort(strings.ToUpper(effort))

	if !converted.IsValid() {
		return model.EffortMedium
	}

	return converted
}

func parseApprovalDecision(decision model.AgentApprovalDecision) string {
	if decision == model.AgentApprovalDecisionAllowOnce {
		return "allow_once"
	}

	return "deny"
}

func convertAgentSessionStatus(status string) model.AgentSessionStatus {
	switch status {
	case agentrunner.StatusWorking:
		return model.AgentSessionStatusWorking
	case agentrunner.StatusWaitingForApproval:
		return model.AgentSessionStatusWaitingForApproval
	default:
		return model.AgentSessionStatusIdle
	}
}

func convertAgentSession(session conductor.AgentSession) model.AgentSession {
	return model.AgentSession{
		ID:        session.ID,
		Title:     session.Title,
		StartedBy: convertWorkspaceMemberOrNil(session.StartedBy),
		StartedAt: session.StartedAt,
		Model:     new(convertAgentModel(session.Model)),
		Effort:    convertEffort(session.Effort),
		Status:    convertAgentSessionStatus(session.Status),
	}
}

func convertWorkspaceMemberOrNil(member *conductor.WorkspaceMember) *model.WorkspaceMember {
	if member == nil {
		return nil
	}

	return convertWorkspaceMember(member)
}

func convertAgentSessionEvent(event conductor.AgentSessionEvent) model.AgentSessionEvent {
	switch event.Type {
	case agentrunner.EventUserMessage:
		return model.AgentUserMessage{ID: event.ID, Time: event.Time, Author: convertWorkspaceMemberOrNil(event.Author), Text: event.Text}
	case agentrunner.EventMessage:
		return model.AgentMessage{ID: event.ID, Time: event.Time, Text: event.Text}
	case agentrunner.EventToolCall:
		call := event.ToolCall

		return model.AgentToolCall{
			ID:        event.ID,
			Time:      event.Time,
			CallID:    call.CallID,
			Connector: optional(call.Connector),
			Tool:      call.Tool,
			Title:     call.Title,
			Status:    model.AgentToolCallStatus(strings.ToUpper(call.Status)),
			Input:     optional(call.Input),
			Output:    optional(call.Output),
		}
	case agentrunner.EventApproval:
		approval := event.Approval

		return model.AgentApproval{
			ID:        event.ID,
			Time:      event.Time,
			Approval:  approval.ApprovalID,
			CallID:    optional(approval.CallID),
			Title:     approval.Title,
			Status:    model.AgentApprovalStatus(strings.ToUpper(approval.Status)),
			DecidedBy: convertWorkspaceMemberOrNil(event.DecidedBy),
		}
	default:
		end := event.TurnEnd

		return model.AgentTurnEnd{
			ID:     event.ID,
			Time:   event.Time,
			Reason: model.AgentTurnEndReason(strings.ToUpper(end.Reason)),
			Error:  optional(end.Error),
		}
	}
}

func parseAgentAuditKinds(kinds []model.AgentAuditKind) []agentaudit.Kind {
	result := make([]agentaudit.Kind, 0, len(kinds))

	for _, kind := range kinds {
		if kind == model.AgentAuditKindModel {
			result = append(result, agentaudit.Model)
		} else {
			result = append(result, agentaudit.Tool)
		}
	}

	return result
}

func convertAgentAuditEntry(entry conductor.AgentAuditEntry) model.AgentAuditEntry {
	result := model.AgentAuditEntry{
		Time:          entry.Time,
		Kind:          model.AgentAuditKindModel,
		ModelProvider: optional(entry.ModelProvider),
		Connector:     optional(entry.Connector),
		Target:        entry.Target,
		InputTokens:   entry.InputTokens,
		OutputTokens:  entry.OutputTokens,
	}

	if entry.Kind == agentaudit.Tool {
		result.Kind = model.AgentAuditKindTool
	}

	switch entry.Result {
	case agentaudit.OK:
		result.Result = model.AgentAuditResultOk
	case agentaudit.Denied:
		result.Result = model.AgentAuditResultDenied
	default:
		result.Result = model.AgentAuditResultFailed
	}

	return result
}
