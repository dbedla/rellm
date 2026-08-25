package rellm

/*
 List of most popular/trending models available for LMS (LM Studio)
 LM Studio does not provide a centralized list of models, so this list is based on user feedback and community recommendations.
 You are not limited to these models. IT is possible to use models outside this list with the rellm library
 In code just simply instead of
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
 Write
		WithModel(rellm.Model("provider/model_name")).
*/

const (
	Model_LMS_OpenAI_Gpt_Oss_20B Model = "openai/gpt-oss-20b"

	Model_LMS_Google_Gemma_4_E4B        Model = "google/gemma-4-e4b"
	Model_LMS_Qwen_Qwen3_5_9B           Model = "qwen/qwen3.5-9b"
	Model_LMS_DeepSeek_R1_0528_Qwen3_8B Model = "deepseek/deepseek-r1-0528-qwen3-8b"

	Model_LMS_Google_Gemma_3_4B  Model = "google/gemma-3-4b"
	Model_LMS_Google_Gemma_3_12B Model = "google/gemma-3-12b"

	//Model_LMS_Google_Gemma_4_26B_A4B Model = "google/gemma-4-26b-a4b"

	Model_LMS_Qwen_Qwen3_Coder_30B Model = "qwen/qwen3-coder-30b"
	Model_LMS_Qwen_Qwen3_6_27B     Model = "qwen/qwen3.6-27b"
	Model_LMS_Qwen_Qwen3_6_35B_A3B Model = "qwen/qwen3.6-35b-a3b"

	Model_LMS_Google_Gemma_4_31B Model = "google/gemma-4-31b"

	Model_LMS_Qwen_Qwen3_5_35B_A3B Model = "qwen/qwen3.5-35b-a3b"
	Model_LMS_Google_Gemma_3N_E4B  Model = "google/gemma-3n-e4b"

	Model_LMS_Qwen_Qwen3_VL_8B            Model = "qwen/qwen3-vl-8b"
	Model_LMS_Qwen_Qwen3_4B_Thinking_2507 Model = "qwen/qwen3-4b-thinking-2507"
	Model_LMS_Qwen_Qwen3_8B               Model = "qwen/qwen3-8b"

	Model_LMS_MistralAI_Ministral_3_14B_Reasoning Model = "mistralai/ministral-3-14b-reasoning"

	Model_LMS_Google_Gemma_4_E2B Model = "google/gemma-4-e2b"

	Model_LMS_ZAI_Org_GLM_4_7_Flash     Model = "zai-org/glm-4.7-flash"
	Model_LMS_NVIDIA_Nemotron_3_Nano_4B Model = "nvidia/nemotron-3-nano-4b"
)
