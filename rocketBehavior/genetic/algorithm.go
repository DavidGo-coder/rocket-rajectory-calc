package genetic

import (
	"math"
	"math/rand/v2"

	"github.com/DavidGo-coder/rocket-trajectory-calc/simulation"
)

func ValidationRocket(pitchDegree, yawDegree, fuel, burnRate, engineEfficiency, budyMass float64) (float64, float64, float64, float64, float64) {
	if pitchDegree < 0 {
		pitchDegree = 0
	}
	if pitchDegree > 90 {
		pitchDegree = 90
	}
	if yawDegree < 0 {
		yawDegree = 0
	}
	if yawDegree > 90 {
		yawDegree = 90
	}
	if fuel < 5.0 {
		fuel = 5.0
	}
	if fuel > budyMass*10.0 {
		fuel = budyMass * 10.0
	}
	if burnRate < 0.5 {
		burnRate = 0.5
	}
	if engineEfficiency < 500.0 {
		engineEfficiency = 500.0
	}
	if engineEfficiency > 2500.0 {
		engineEfficiency = 2500.0
	}
	return pitchDegree, yawDegree, fuel, burnRate, engineEfficiency
}

func FirstGeneticCalculation(firstRocket simulation.Rocket, populationSize int, rocketBudyMass float64) []simulation.Rocket {
	rocketPopulation := make([]simulation.Rocket, 0, populationSize)

	firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency = ValidationRocket(firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency, rocketBudyMass)

	rocketPopulation = append(rocketPopulation, firstRocket)

	if populationSize > 5 {
		rocketPopulation = append(rocketPopulation, simulation.Rocket{PitchDegree: 45.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000})

		rocketPopulation = append(rocketPopulation, simulation.Rocket{PitchDegree: 35.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: 8.0, EngineEfficiency: 2000})

		rocketPopulation = append(rocketPopulation, simulation.Rocket{PitchDegree: 65.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000})
	}

	for i := len(rocketPopulation); i < populationSize; i++ {
		pitchRand := firstRocket.PitchDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		yawRand := firstRocket.YawDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		fuelRand := firstRocket.Fuel + (((rand.Float64() * 2.0) - 1.0) * (firstRocket.Fuel * 0.1))
		burnRateRand := firstRocket.BurnRate + (((rand.Float64() * 2.0) - 1.0) * 0.5)
		engineEfficiencyRand := firstRocket.EngineEfficiency + ((rand.Float64()*2.0 - 1.0) * 400.0)

		pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand = ValidationRocket(pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand, rocketBudyMass)

		RocketRand := simulation.Rocket{PitchDegree: pitchRand, YawDegree: yawRand, Fuel: fuelRand, BurnRate: burnRateRand, EngineEfficiency: engineEfficiencyRand}
		rocketPopulation = append(rocketPopulation, RocketRand)
	}
	return rocketPopulation
}

func GeneticCalculation(sliceTenBestRocket []simulation.Rocket, populationSize int, medianPopulation, rocketBudyMass float64) []simulation.Rocket {
	rocketPopulation2 := make([]simulation.Rocket, 0, populationSize)

	lenSliceBestRocket := len(sliceTenBestRocket)
	if lenSliceBestRocket == 0 {
		return make([]simulation.Rocket, populationSize)
	}

	for i := 0; i < lenSliceBestRocket && len(rocketPopulation2) < populationSize; i++ {
		rocketPopulation2 = append(rocketPopulation2, sliceTenBestRocket[i])
	}

	sliceTenBestSortedRocket := make([]simulation.Rocket, lenSliceBestRocket)
	for a := 0; a < lenSliceBestRocket; a++ {
		randInt := rand.IntN(lenSliceBestRocket)
		sliceTenBestSortedRocket[a] = sliceTenBestRocket[randInt]
	}

	parentIndex := 0
	for len(rocketPopulation2) < populationSize {
		firstParent := sliceTenBestSortedRocket[parentIndex%lenSliceBestRocket]
		secondParent := sliceTenBestSortedRocket[(parentIndex+1)%lenSliceBestRocket]
		parentIndex++

		if len(rocketPopulation2)%5 == 0 {
			exclusivePitchRand := rand.Float64()
			exclusiveYawRand := rand.Float64()
			exclusiveFuelRand := rand.Float64()
			exclusiveBurnRateRand := rand.Float64()
			exclusiveEngineEfficiencyRand := rand.Float64()

			exclusivePitch := (firstParent.PitchDegree*exclusivePitchRand + secondParent.PitchDegree*(1.0-exclusivePitchRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			exclusiveYaw := (firstParent.YawDegree*exclusiveYawRand + secondParent.YawDegree*(1.0-exclusiveYawRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			exclusiveFuel := (firstParent.Fuel*exclusiveFuelRand+secondParent.Fuel*(1.0-exclusiveFuelRand))*(medianPopulation/500+1.0) + rand.Float64()*4.0 - 2.0
			exclusiveBurnRate := math.Max(0.5, (firstParent.BurnRate*exclusiveBurnRateRand+secondParent.BurnRate*(1.0-exclusiveBurnRateRand))*(medianPopulation/500+1.0)+rand.Float64()*0.4-0.2)
			exclusiveEngineEfficiency := (firstParent.EngineEfficiency*exclusiveEngineEfficiencyRand+secondParent.EngineEfficiency*(1.0-exclusiveEngineEfficiencyRand))*(medianPopulation/500+1.0) + rand.Float64()*4.0 - 2.0

			exclusivePitch, exclusiveYaw, exclusiveFuel, exclusiveBurnRate, exclusiveEngineEfficiency = ValidationRocket(exclusivePitch, exclusiveYaw, exclusiveFuel, exclusiveBurnRate, exclusiveEngineEfficiency, rocketBudyMass)

			exclusiveChild := simulation.Rocket{
				PitchDegree:      exclusivePitch,
				YawDegree:        exclusiveYaw,
				Fuel:             exclusiveFuel,
				BurnRate:         exclusiveBurnRate,
				EngineEfficiency: exclusiveEngineEfficiency,
			}
			rocketPopulation2 = append(rocketPopulation2, exclusiveChild)

			if len(rocketPopulation2) == populationSize {
				break
			}
		}

		childPitchRand := rand.Float64()
		childYawRand := rand.Float64()
		childFuelRand := rand.Float64()
		childBurnRateRand := rand.Float64()
		childEngineEfficiencyRand := rand.Float64()

		childPitch := (firstParent.PitchDegree*childPitchRand + secondParent.PitchDegree*(1.0-childPitchRand)) + rand.Float64()*6.0 - 3.0
		childYaw := (firstParent.YawDegree*childYawRand + secondParent.YawDegree*(1.0-childYawRand)) + rand.Float64()*6.0 - 3.0
		childFuel := (firstParent.Fuel*childFuelRand + secondParent.Fuel*(1.0-childFuelRand)) + rand.Float64()*4.0 - 2.0
		childBurnRate := math.Max(0.5, (firstParent.BurnRate*childBurnRateRand+secondParent.BurnRate*(1.0-childBurnRateRand))+rand.Float64()*0.4-0.2)
		childEngineEfficiency := (firstParent.EngineEfficiency*childEngineEfficiencyRand + secondParent.EngineEfficiency*(1.0-childEngineEfficiencyRand)) + rand.Float64()*10.0 - 5.0

		childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency = ValidationRocket(childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency, rocketBudyMass)

		child := simulation.Rocket{
			PitchDegree:      childPitch,
			YawDegree:        childYaw,
			Fuel:             childFuel,
			BurnRate:         childBurnRate,
			EngineEfficiency: childEngineEfficiency,
		}
		rocketPopulation2 = append(rocketPopulation2, child)
	}

	return rocketPopulation2
}
