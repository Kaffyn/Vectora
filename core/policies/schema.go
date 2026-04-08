package policies

// Policy represents the baseline structure of an unnegotiable system rule.
type Policy struct {
	PolicyID    string      `yaml:"policy_id"`
	Rule        string      `yaml:"rule"`
	Enforcement Enforcement `yaml:"enforcement"`
}

// Enforcement maps all enforcement actions. Irrelevant fields across YAML files will be ignored.
type Enforcement struct {
	// Scope Guardian Actions
	PreCondition    string   `yaml:"pre_condition,omitempty"`
	OnViolation     string   `yaml:"on_violation,omitempty"`
	BlockExtensions []string `yaml:"block_extensions,omitempty"`
	BlockFiles      []string `yaml:"block_files,omitempty"`

	// Passive Git Actions
	PassiveCheck    string `yaml:"passive_check,omitempty"`
	NoForcedInstall string `yaml:"no_forced_install,omitempty"`
	NoForcedInit    string `yaml:"no_forced_init,omitempty"`

	// Snapshots and Auth Gates
	PreWriteAction  string `yaml:"pre_write_action,omitempty"`
	ActionStrategy  string `yaml:"action_strategy,omitempty"`
	CommitMessaging string `yaml:"commit_messaging,omitempty"`
	ViaAcpIde       string `yaml:"via_acp_ide,omitempty"`
	ViaMcpCli       string `yaml:"via_mcp_cli,omitempty"`

	// RAG Priority rules
	PromptInjection    string      `yaml:"prompt_injection,omitempty"`
	BlockIfNoRetrieval interface{} `yaml:"block_if_no_retrieval,omitempty"`
	FallbackAction     string      `yaml:"fallback_action,omitempty"`
}

// PolicyEngine maintains loaded policies in memory
type PolicyEngine struct {
	TrustFolder   string
	ScopePolicies []Policy
	GitPolicies   []Policy
	RAGPolicies   []Policy
}
