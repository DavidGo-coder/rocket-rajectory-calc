package simulation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

func CalculationTrajectory(targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize int, rocketPopulation []Rocket) ([]Rocket, float64) {
	var wg sync.WaitGroup
	wg.Add(populationSize)

	type ResultSimulation struct {
		x             float64
		y             float64
		trajectoryCSV string
		score         float64
		resultIndex   int
	}

	channelResultSimulation := make(chan ResultSimulation, populationSize)

	for countSimulation := 0; countSimulation < populationSize; countSimulation++ {
		go func(index int, fuel, pitchDegree, yawDegree, burnRate, engineEfficiency, bodyMass float64) {
			defer wg.Done()

			pitchRadians := pitchDegree * math.Pi / 180.0
			vx, vy := 0.0, 0.0
			x, y := 0.0, 0.0
			time := 0.0
			maxY := 0.0

			var trajectoryPoints []string
			stepCounter := 0

			targetX := float64(targetCoordinateX)
			targetY := float64(targetCoordinateY)

			minDistToTarget := math.Sqrt(targetX*targetX + targetY*targetY)
			currentRocket := rocketPopulation[index]

			for time < 275.0 {
				currentMass := currentRocket.BodyMass + fuel
				if currentMass < 0.1 {
					currentMass = 0.1
				}

				var thrustForceX, thrustForceY float64
				if fuel > 0 {
					dynamicThrust := burnRate * engineEfficiency

					fuelConsumed := burnRate * OneMomentSimulation
					if fuelConsumed > fuel {
						dynamicThrust = (fuel / OneMomentSimulation) * engineEfficiency
						fuel = 0
					} else {
						fuel -= fuelConsumed
					}

					thrustForceX = dynamicThrust * math.Cos(pitchRadians)
					thrustForceY = dynamicThrust * math.Sin(pitchRadians)
				}

				var fdragX, fdragY float64
				v := math.Sqrt(vx*vx + vy*vy)
				if v > 0.001 {
					currentDragCoef := AirResistance * math.Exp(-y/8500.0)
					fdragTotal := currentDragCoef * v * v

					fdragX = fdragTotal * (vx / v)
					fdragY = fdragTotal * (vy / v)
				}

				t := RadiusPlanetEarth / (RadiusPlanetEarth + y)
				currentGravitation := GravitationalConstant * (t * t)

				ax := (thrustForceX - fdragX) / currentMass
				ay := ((thrustForceY - fdragY) / currentMass) - currentGravitation

				x += vx * OneMomentSimulation
				y += vy * OneMomentSimulation
				vx += ax * OneMomentSimulation
				vy += ay * OneMomentSimulation

				if y <= 0 && vy < 0 {
					y = 0

					if time < 0.5 || thrustForceY > currentGravitation {
						vy = 0

					} else {

						x -= vx * OneMomentSimulation
						vx = 0
						vy = 0

						break
					}
				}

				if y > maxY {
					maxY = y
				}

				dx := targetX - x
				dy := targetY - y
				currentDist := math.Sqrt(dx*dx + dy*dy)

				if currentDist < minDistToTarget {
					minDistToTarget = currentDist
				}

				if stepCounter%50 == 0 {
					trajectoryPoints = append(trajectoryPoints, fmt.Sprintf("%.2f,%.2f", x, y))
				}
				stepCounter++
				time += OneMomentSimulation
			}

			trajectoryPoints = append(trajectoryPoints, fmt.Sprintf("%.2f,%.2f", x, y))

			efficiencyScore := minDistToTarget

			if maxY < 5.0 {
				efficiencyScore += FailedLaunchPenalty // potentional weak area with "magic" number
			} else {
				proximity := math.Exp(-minDistToTarget / 500.0) // also magic number "500.0"
				efficiencyScore -= fuel * 2.0 * proximity
			}

			channelResultSimulation <- ResultSimulation{
				x:             x,
				y:             maxY,
				score:         efficiencyScore,
				trajectoryCSV: strings.Join(trajectoryPoints, ";"),
				resultIndex:   index,
			}
		}(countSimulation, rocketPopulation[countSimulation].Fuel, rocketPopulation[countSimulation].PitchDegree, rocketPopulation[countSimulation].YawDegree, rocketPopulation[countSimulation].BurnRate, rocketPopulation[countSimulation].EngineEfficiency, rocketPopulation[countSimulation].BodyMass)
	}
	wg.Wait()
	close(channelResultSimulation)

	for result := range channelResultSimulation {
		rocketPopulation[result.resultIndex].Fitness = result.score
		rocketPopulation[result.resultIndex].X = result.x
		rocketPopulation[result.resultIndex].Y = result.y
		rocketPopulation[result.resultIndex].TrajectoryCSV = result.trajectoryCSV
	}

	sort.Slice(rocketPopulation[:populationSize], func(i, j int) bool {
		return rocketPopulation[i].Fitness < rocketPopulation[j].Fitness
	})

	medianFitness := rocketPopulation[populationSize/2].Fitness
	fmt.Printf("🧬 [EVOLUTION] Processing generation... Median target miss: %.2f meters\n", medianFitness)

	var sliceTenBestRocket []Rocket
	for b := 0; b < 10 && b < populationSize; b++ {
		sliceTenBestRocket = append(sliceTenBestRocket, rocketPopulation[b])
	}

	return sliceTenBestRocket, medianFitness
}
