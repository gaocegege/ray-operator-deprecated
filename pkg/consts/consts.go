package consts

const (
	EventNormal  = "Normal"
	EventWarning = "Warning"

	ReasonValidationFailed = "ValidationFailedOrNotImplemented"
	ReasonCreate           = "SuccessfullyCreate"
	ReasonUpdate           = "SuccessfullyUpdate"

	LabelRayWorker = "ray-worker"
	LabelRayHead   = "ray-head"
	LabelRay       = "ray"

	EnvNodeIP         = "RAY_NODE_IP"
	FieldPathPodIP    = "status.podIP"
	EnvRayHeadService = "RAY_HEAD_SERVICE"

	ContainerRayHead = "ray-head"
)
