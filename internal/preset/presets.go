package preset

// SearchPreset defines a predefined search configuration.
type SearchPreset struct {
	Name        string   // Preset name
	Description string   // What this preset searches for
	Keywords    []string // Keywords to combine
	Query       string   // Final combined query
	MinScore    int      // Recommended minimum score
	MaxAgeDays  int      // Recommended max age in days
}

// Presets contains all available search presets.
var Presets = map[string]SearchPreset{
	// LLM & NLP
	"llm-reasoning": {
		Name:        "llm-reasoning",
		Description: "LLM reasoning and chain-of-thought",
		Keywords:    []string{"large language model", "reasoning", "chain of thought", "CoT"},
		Query:       "large language model reasoning chain of thought",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"llm-agent": {
		Name:        "llm-agent",
		Description: "LLM-based agents and tool use",
		Keywords:    []string{"large language model", "agent", "tool use", "planning"},
		Query:       "large language model agent tool use planning",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"llm-eval": {
		Name:        "llm-eval",
		Description: "LLM evaluation and benchmarks",
		Keywords:    []string{"large language model", "evaluation", "benchmark", "assessment"},
		Query:       "large language model evaluation benchmark",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"rag": {
		Name:        "rag",
		Description: "Retrieval-Augmented Generation",
		Keywords:    []string{"retrieval augmented generation", "RAG", "knowledge retrieval"},
		Query:       "retrieval augmented generation RAG",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"prompt": {
		Name:        "prompt",
		Description: "Prompt engineering and optimization",
		Keywords:    []string{"prompt engineering", "prompt optimization", "in-context learning"},
		Query:       "prompt engineering optimization in-context learning",
		MinScore:    50,
		MaxAgeDays:  180,
	},

	// Computer Vision
	"diffusion": {
		Name:        "diffusion",
		Description: "Diffusion models for image generation",
		Keywords:    []string{"diffusion model", "image generation", "stable diffusion"},
		Query:       "diffusion model image generation",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"multimodal": {
		Name:        "multimodal",
		Description: "Multimodal learning and vision-language",
		Keywords:    []string{"multimodal", "vision language", "CLIP", "visual understanding"},
		Query:       "multimodal vision language model",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"video": {
		Name:        "video",
		Description: "Video understanding and generation",
		Keywords:    []string{"video understanding", "video generation", "temporal modeling"},
		Query:       "video understanding generation temporal",
		MinScore:    50,
		MaxAgeDays:  180,
	},

	// Machine Learning
	"transformer": {
		Name:        "transformer",
		Description: "Transformer architecture improvements",
		Keywords:    []string{"transformer", "attention mechanism", "efficient transformer"},
		Query:       "transformer attention mechanism efficient",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"finetune": {
		Name:        "finetune",
		Description: "Fine-tuning and adaptation methods",
		Keywords:    []string{"fine-tuning", "LoRA", "adapter", "parameter efficient"},
		Query:       "fine-tuning LoRA adapter parameter efficient",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"distill": {
		Name:        "distill",
		Description: "Knowledge distillation and compression",
		Keywords:    []string{"knowledge distillation", "model compression", "pruning", "quantization"},
		Query:       "knowledge distillation model compression",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"rl": {
		Name:        "rl",
		Description: "Reinforcement learning",
		Keywords:    []string{"reinforcement learning", "RLHF", "reward model", "policy optimization"},
		Query:       "reinforcement learning RLHF reward model",
		MinScore:    50,
		MaxAgeDays:  180,
	},

	// Safety & Alignment
	"alignment": {
		Name:        "alignment",
		Description: "AI alignment and safety",
		Keywords:    []string{"AI alignment", "safety", "value alignment", "constitutional AI"},
		Query:       "AI alignment safety value",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"jailbreak": {
		Name:        "jailbreak",
		Description: "Jailbreak attacks and defenses",
		Keywords:    []string{"jailbreak", "adversarial attack", "LLM security", "red teaming"},
		Query:       "jailbreak adversarial attack LLM security",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"hallucination": {
		Name:        "hallucination",
		Description: "Hallucination detection and mitigation",
		Keywords:    []string{"hallucination", "factuality", "faithfulness", "grounding"},
		Query:       "hallucination detection factuality LLM",
		MinScore:    50,
		MaxAgeDays:  180,
	},

	// Data & Training
	"data-synthesis": {
		Name:        "data-synthesis",
		Description: "Synthetic data generation",
		Keywords:    []string{"synthetic data", "data augmentation", "data generation"},
		Query:       "synthetic data generation augmentation",
		MinScore:    50,
		MaxAgeDays:  180,
	},
	"scaling": {
		Name:        "scaling",
		Description: "Scaling laws and large-scale training",
		Keywords:    []string{"scaling law", "large scale training", "compute optimal"},
		Query:       "scaling law large scale training",
		MinScore:    50,
		MaxAgeDays:  180,
	},
}

// List returns all preset names and descriptions.
func List() []SearchPreset {
	result := make([]SearchPreset, 0, len(Presets))
	for _, p := range Presets {
		result = append(result, p)
	}
	return result
}

// Get returns a preset by name.
func Get(name string) (SearchPreset, bool) {
	p, ok := Presets[name]
	return p, ok
}
