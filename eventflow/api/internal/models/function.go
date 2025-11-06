package models


import "time"

type Function struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Command   []string          `json:"command,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Replicas  int32             `json:"replicas"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type FunctionStatus struct {
	Name              string    `json:"name"`
	Image             string    `json:"image"`
	Replicas          int32     `json:"replicas"`
	AvailableReplicas int32     `json:"available_replicas"`
	ReadyReplicas     int32     `json:"ready_replicas"`
	UpdatedReplicas   int32     `json:"updated_replicas"`
	Status            string    `json:"status"` // Running, Pending, Failed
	CreatedAt         time.Time `json:"created_at"`
}

type CreateFunctionRequest struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Command  []string          `json:"command,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Replicas int32             `json:"replicas"`
}

type InvokeFunctionRequest struct {
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
