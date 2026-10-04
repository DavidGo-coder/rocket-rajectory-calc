package simulation

type Rocket struct {
	PitchDegree      float64
	YawDegree        float64
	Fuel             float64
	BurnRate         float64
	BodyMass         float64
	EngineEfficiency float64
	Fitness          float64
	X, Y, Z          float64
	TrajectoryCSV    string
}

const (
	ThrustForce           = 1500.0
	OneMomentSimulation   = 0.01
	GravitationalConstant = 9.81
	RadiusPlanetEarth     = 6371000.0
	AirResistance         = 0.02
	FailedLaunchPenalty   = 50000.0
)
