package common

// TODO: fill in the details.

type GRpcImposter struct {
	Request  GRpcRequest  `json:"request"`
	Response GRpcResponse `json:"response"`
}

type GRpcRequest struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type GRpcResponse struct {
	Headers map[string]string `json:"headers"`
}
