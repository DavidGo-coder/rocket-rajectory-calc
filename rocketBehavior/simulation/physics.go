package simulation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

func CalculationTrajectory(TargetCoordinateX, TargetCoordinateY, TargetCoordinateZ, PopulationSize int, BodyMassPopulation float64, RocketPopulation2 []Rocket) ([]Rocket, float64) {
	var wg sync.WaitGroup
	wg.Add(PopulationSize)

	type ResultSimulation struct {
		X             float64
		Y             float64
		TrajectoryCSV string
		Sucess        float64
		ResultIndex   int
	}

	ChannelResultSimulation := make(chan ResultSimulation, PopulationSize)

	for CountPopulationSizeSimulation := 0; CountPopulationSizeSimulation < PopulationSize; CountPopulationSizeSimulation++ {
		go func(Index int, Fuel, PitchDegree, YawDegree, BurnRate, EngineEfficiency float64) {
			defer wg.Done()

			PitchRadians := PitchDegree * math.Pi / 180.0
			Vx, Vy := 0.0, 0.0
			X, Y := 0.0, 0.0
			Time := 0.0
			maxY := 0.0

			var TrajectoryPoints []string
			StepCounter := 0

			targetX := float64(TargetCoordinateX)
			targetY := float64(TargetCoordinateY)

			MinDistToTarget := math.Sqrt(targetX*targetX + targetY*targetY)

			for Time < 150.0 {
				CurrentMass := BodyMassPopulation + Fuel
				if CurrentMass < BodyMassPopulation {
					CurrentMass = BodyMassPopulation
				}

				var ThrustForceX, ThrustForceY float64
				if Fuel > 0 {
					DynamicThrust := BurnRate * EngineEfficiency

					FuelConsumed := BurnRate * OneMomentSimulation
					if FuelConsumed > Fuel {
						DynamicThrust = (Fuel / OneMomentSimulation) * EngineEfficiency
						Fuel = 0
					} else {
						Fuel -= FuelConsumed
					}

					ThrustForceX = DynamicThrust * math.Cos(PitchRadians)
					ThrustForceY = DynamicThrust * math.Sin(PitchRadians)
				}

				var FdragX, FdragY float64
				V := math.Sqrt(Vx*Vx + Vy*Vy)
				if V > 0.001 {
					CurrentDragCoef := AirResistance * math.Exp(-Y/8500.0)
					FdragTotal := CurrentDragCoef * V * V

					FdragX = FdragTotal * (Vx / V)
					FdragY = FdragTotal * (Vy / V)
				}

				CurrentGravitation := GravitationalConstant * math.Pow(RadiusPlanetEarth/(RadiusPlanetEarth+Y), 2)

				Ax := (ThrustForceX - FdragX) / CurrentMass
				Ay := ((ThrustForceY - FdragY) / CurrentMass) - CurrentGravitation

				X += Vx * OneMomentSimulation
				Y += Vy * OneMomentSimulation
				Vx += Ax * OneMomentSimulation
				Vy += Ay * OneMomentSimulation

				if Y <= 0 && Vy < 0 {
					Y = 0
					Vx = 0
					Vy = 0
					if Time > 0.5 {
						break
					}
				}

				if Y > maxY {
					maxY = Y
				}

				Dx := targetX - X
				Dy := targetY - Y
				CurrentDist := math.Sqrt(Dx*Dx + Dy*Dy)

				if CurrentDist < MinDistToTarget {
					MinDistToTarget = CurrentDist
				}

				if StepCounter%50 == 0 {
					TrajectoryPoints = append(TrajectoryPoints, fmt.Sprintf("%.2f,%.2f", X, Y))
				}
				StepCounter++
				Time += OneMomentSimulation
			}

			TrajectoryPoints = append(TrajectoryPoints, fmt.Sprintf("%.2f,%.2f", X, Y))

			EfficiencyRating := MinDistToTarget

			if X < 5.0 && maxY < 5.0 {
				EfficiencyRating += 50000.0
			} else {

				if MinDistToTarget < 500.0 {
					EfficiencyRating -= Fuel * 2.0
				}
			}

			ChannelResultSimulation <- ResultSimulation{
				X:             X,
				Y:             maxY,
				Sucess:        EfficiencyRating,
				TrajectoryCSV: strings.Join(TrajectoryPoints, ";"),
				ResultIndex:   Index,
			}
		}(CountPopulationSizeSimulation, RocketPopulation2[CountPopulationSizeSimulation].Fuel, RocketPopulation2[CountPopulationSizeSimulation].PitchDegree, RocketPopulation2[CountPopulationSizeSimulation].YawDegree, RocketPopulation2[CountPopulationSizeSimulation].BurnRate, RocketPopulation2[CountPopulationSizeSimulation].EngineEfficiency)
	}
	wg.Wait()
	close(ChannelResultSimulation)

	for Result := range ChannelResultSimulation {
		RocketPopulation2[Result.ResultIndex].Fitness = Result.Sucess
		RocketPopulation2[Result.ResultIndex].X = Result.X
		RocketPopulation2[Result.ResultIndex].Y = Result.Y
		RocketPopulation2[Result.ResultIndex].TrajectoryCSV = Result.TrajectoryCSV
	}

	sort.Slice(RocketPopulation2[:PopulationSize], func(i, j int) bool {
		return RocketPopulation2[i].Fitness < RocketPopulation2[j].Fitness
	})

	MedianPopulation := RocketPopulation2[PopulationSize/2].Fitness
	fmt.Printf("🧬 [EVOLUTION] Processing generation... Median target miss: %.2f meters\n", MedianPopulation)

	var SliceTenBestRocket []Rocket
	for b := 0; b < 10 && b < PopulationSize; b++ {
		SliceTenBestRocket = append(SliceTenBestRocket, RocketPopulation2[b])
	}

	return SliceTenBestRocket, MedianPopulation
}
