package util

import "encoding/json"

// RewriteChatCompletionModel rewrites the `model` field in request body
// if modelMap contains a mapping for the original model.
// Only applies to top-level "model" field.
func RewriteChatCompletionModel(
	requestBody []byte,
	modelMap map[string]string,
) ([]byte, bool, error) {

	if len(modelMap) == 0 {
		return requestBody, false, nil
	}

	var body map[string]interface{}
	if err := json.Unmarshal(requestBody, &body); err != nil {
		return requestBody, false, err
	}

	rawModel, ok := body["model"]
	if !ok {
		return requestBody, false, nil
	}

	model, ok := rawModel.(string)
	if !ok {
		return requestBody, false, nil
	}

	newModel, ok := modelMap[model]
	if !ok || newModel == "" || newModel == model {
		return requestBody, false, nil
	}

	// 🔒 only rewrite this field
	body["model"] = newModel

	newBody, err := json.Marshal(body)
	if err != nil {
		return requestBody, false, err
	}

	return newBody, true, nil
}
