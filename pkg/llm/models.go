package llm

import "fmt"

type Model string

const (
	ModelAllam27B                      Model = "allam-2-7b"
	ModelGroqCompound                  Model = "groq/compound"
	ModelGroqCompoundMini              Model = "groq/compound-mini"
	ModelLlama318BInstant              Model = "llama-3.1-8b-instant"
	ModelLlama3370BVersatile           Model = "llama-3.3-70b-versatile"
	ModelLlama4Maverick17B128EInstruct Model = "meta-llama/llama-4-maverick-17b-128e-instruct"
	ModelLlama4Scout17B16EInstruct     Model = "meta-llama/llama-4-scout-17b-16e-instruct"
	ModelLlamaGuard412B                Model = "meta-llama/llama-guard-4-12b"
	ModelLlamaPromptGuard222M          Model = "meta-llama/llama-prompt-guard-2-22m"
	ModelLlamaPromptGuard286M          Model = "meta-llama/llama-prompt-guard-2-86m"
	ModelKimiK2Instruct                Model = "moonshotai/kimi-k2-instruct"
	ModelKimiK2Instruct0905            Model = "moonshotai/kimi-k2-instruct-0905"
	ModelOpenAIGptOss120B              Model = "openai/gpt-oss-120b"
	ModelOpenAIGptOss20B               Model = "openai/gpt-oss-20b"
	ModelOpenAIGptOssSafeguard20B      Model = "openai/gpt-oss-safeguard-20b"
	ModelQwen32B                       Model = "qwen/qwen3-32b"
)

func (m *Model) String() string {
	return fmt.Sprintf("%v", *m)
}
func (m *Model) Set(value string) error {
	*m = Model(value)
	return nil
}

var modelSet = map[Model]struct{}{
	ModelAllam27B:                      {},
	ModelGroqCompound:                  {},
	ModelGroqCompoundMini:              {},
	ModelLlama318BInstant:              {},
	ModelLlama3370BVersatile:           {},
	ModelLlama4Maverick17B128EInstruct: {},
	ModelLlama4Scout17B16EInstruct:     {},
	ModelLlamaGuard412B:                {},
	ModelLlamaPromptGuard222M:          {},
	ModelLlamaPromptGuard286M:          {},
	ModelKimiK2Instruct:                {},
	ModelKimiK2Instruct0905:            {},
	ModelOpenAIGptOss120B:              {},
	ModelOpenAIGptOss20B:               {},
	ModelOpenAIGptOssSafeguard20B:      {},
	ModelQwen32B:                       {},
}

func (m Model) IsValid() bool {
	_, ok := modelSet[m]
	return ok
}

func All() []Model {
	models := make([]Model, 0, len(modelSet))
	for m := range modelSet {
		models = append(models, m)
	}
	return models
}
