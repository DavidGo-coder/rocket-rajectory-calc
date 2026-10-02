package simulation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

func CalculationTrajectory(targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize int, bodyMassPopulation float64, rocketPopulation2 []Rocket) ([]Rocket, float64) {
	var wg sync.WaitGroup
	wg.Add(populationSize)

	type ResultSimulation struct {
		x             float64
		y             float64
		trajectoryCSV string
		sucess        float64
		resultIndex   int
	}

	channelResultSimulation := make(chan ResultSimulation, populationSize)

	for countPopulationSizeSimulation := 0; countPopulationSizeSimulation < populationSize; countPopulationSizeSimulation++ {
		go func(index int, fuel, pitchDegree, yawDegree, burnRate, engineEfficiency float64) {
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

			for time < 150.0 {
				currentMass := bodyMassPopulation + fuel
				if currentMass < bodyMassPopulation {
					currentMass = bodyMassPopulation
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

				currentGravitation := GravitationalConstant * math.Pow(RadiusPlanetEarth/(RadiusPlanetEarth+y), 2)

				ax := (thrustForceX - fdragX) / currentMass
				ay := ((thrustForceY - fdragY) / currentMass) - currentGravitation

				x += vx * OneMomentSimulation
				y += vy * OneMomentSimulation
				vx += ax * OneMomentSimulation
				vy += ay * OneMomentSimulation

				if y <= 0 && vy < 0 {
					y = 0
					vx = 0
					vy = 0
					if time > 0.5 {
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

			efficiencyRating := minDistToTarget

			if x < 5.0 && maxY < 5.0 {
				efficiencyRating += 50000.0
			} else {

				if minDistToTarget < 500.0 {
					efficiencyRating -= fuel * 2.0
				}
			}

			channelResultSimulation <- ResultSimulation{
				x:             x,
				y:             maxY,
				sucess:        efficiencyRating,
				trajectoryCSV: strings.Join(trajectoryPoints, ";"),
				resultIndex:   index,
			}
		}(countPopulationSizeSimulation, rocketPopulation2[countPopulationSizeSimulation].Fuel, rocketPopulation2[countPopulationSizeSimulation].PitchDegree, rocketPopulation2[countPopulationSizeSimulation].YawDegree, rocketPopulation2[countPopulationSizeSimulation].BurnRate, rocketPopulation2[countPopulationSizeSimulation].EngineEfficiency)
	}
	wg.Wait()
	close(channelResultSimulation)

	for result := range channelResultSimulation {
		rocketPopulation2[result.resultIndex].Fitness = result.sucess
		rocketPopulation2[result.resultIndex].X = result.x
		rocketPopulation2[result.resultIndex].Y = result.y
		rocketPopulation2[result.resultIndex].TrajectoryCSV = result.trajectoryCSV
	}

	sort.Slice(rocketPopulation2[:populationSize], func(i, j int) bool {
		return rocketPopulation2[i].Fitness < rocketPopulation2[j].Fitness
	})

	medianPopulation := rocketPopulation2[populationSize/2].Fitness
	fmt.Printf("🧬 [EVOLUTION] Processing generation... Median target miss: %.2f meters\n", medianPopulation)

	var sliceTenBestRocket []Rocket
	for b := 0; b < 10 && b < populationSize; b++ {
		sliceTenBestRocket = append(sliceTenBestRocket, rocketPopulation2[b])
	}

	return sliceTenBestRocket, medianPopulation
}
