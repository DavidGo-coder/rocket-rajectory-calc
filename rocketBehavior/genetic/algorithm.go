package genetic

import (
	"math"
	"math/rand/v2"

	"github.com/DavidGo-coder/rocket-trajectory-calc/simulation"
)

func validationRocket(pitchDegree, yawDegree, fuel, burnRate, engineEfficiency, bodyMass float64) (float64, float64, float64, float64, float64) {
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
	if fuel > bodyMass*10.0 {
		fuel = bodyMass * 10.0
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

func FirstGeneticCalculation(firstRocket simulation.Rocket, populationSize int) []simulation.Rocket {
	initialPopulation := make([]simulation.Rocket, 0, populationSize)

	firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency = validationRocket(firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency, firstRocket.BodyMass)

	initialPopulation = append(initialPopulation, firstRocket)

	if populationSize > 5 {
		initialPopulation = append(initialPopulation, simulation.Rocket{PitchDegree: 45.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000})

		initialPopulation = append(initialPopulation, simulation.Rocket{PitchDegree: 35.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: 8.0, EngineEfficiency: 2000})

		initialPopulation = append(initialPopulation, simulation.Rocket{PitchDegree: 65.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000})
	}

	for i := len(initialPopulation); i < populationSize; i++ {
		pitchRand := firstRocket.PitchDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		yawRand := firstRocket.YawDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		fuelRand := firstRocket.Fuel + (((rand.Float64() * 2.0) - 1.0) * (firstRocket.Fuel * 0.1))
		burnRateRand := firstRocket.BurnRate + (((rand.Float64() * 2.0) - 1.0) * 0.5)
		engineEfficiencyRand := firstRocket.EngineEfficiency + ((rand.Float64()*10.0 - 5.0) * 400.0)

		pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand = validationRocket(pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand, firstRocket.BodyMass)

		rocketRand := simulation.Rocket{
			PitchDegree:      pitchRand,
			YawDegree:        yawRand,
			Fuel:             fuelRand,
			BurnRate:         burnRateRand,
			EngineEfficiency: engineEfficiencyRand,
			BodyMass:         firstRocket.BodyMass,
		}
		initialPopulation = append(initialPopulation, rocketRand)
	}
	return initialPopulation
}

func GeneticCalculation(sliceTenBestRocket []simulation.Rocket, populationSize int, medianFitness, bodyMass float64) []simulation.Rocket {
	nextRocketPopulation := make([]simulation.Rocket, 0, populationSize)

	length := len(sliceTenBestRocket)
	if length == 0 {
		return make([]simulation.Rocket, populationSize)
	}

	for i := 0; i < length && len(nextRocketPopulation) < populationSize; i++ {
		nextRocketPopulation = append(nextRocketPopulation, sliceTenBestRocket[i])
	}

	sliceTenBestShuffledRocket := make([]simulation.Rocket, length)
	copy(sliceTenBestShuffledRocket, sliceTenBestRocket)

	rand.Shuffle(length, func(i, j int) {
		sliceTenBestShuffledRocket[i], sliceTenBestShuffledRocket[j] = sliceTenBestShuffledRocket[j], sliceTenBestShuffledRocket[i]
	})

	parentIndex := 0
	for len(nextRocketPopulation) < populationSize {
		firstParent := sliceTenBestShuffledRocket[parentIndex%length]
		secondParent := sliceTenBestShuffledRocket[(parentIndex+1)%length]
		parentIndex++

		if len(nextRocketPopulation)%5 == 0 {
			boostedPitchRand := rand.Float64()
			boostedYawRand := rand.Float64()
			boostedFuelRand := rand.Float64()
			boostedBurnRateRand := rand.Float64()
			boostedEngineEfficiencyRand := rand.Float64()

			boostedPitch := (firstParent.PitchDegree*boostedPitchRand + secondParent.PitchDegree*(1.0-boostedPitchRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			boostedYaw := (firstParent.YawDegree*boostedYawRand + secondParent.YawDegree*(1.0-boostedYawRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			boostedFuel := (firstParent.Fuel*boostedFuelRand+secondParent.Fuel*(1.0-boostedFuelRand))*(medianFitness/500+1.0) + rand.Float64()*4.0 - 2.0
			boostedBurnRate := math.Max(0.5, (firstParent.BurnRate*boostedBurnRateRand+secondParent.BurnRate*(1.0-boostedBurnRateRand))*(medianFitness/500+1.0)+rand.Float64()*0.4-0.2)
			boostedEngineEfficiency := (firstParent.EngineEfficiency*boostedEngineEfficiencyRand+secondParent.EngineEfficiency*(1.0-boostedEngineEfficiencyRand))*(medianFitness/500+1.0) + rand.Float64()*10.0 - 5.0

			boostedPitch, boostedYaw, boostedFuel, boostedBurnRate, boostedEngineEfficiency = validationRocket(boostedPitch, boostedYaw, boostedFuel, boostedBurnRate, boostedEngineEfficiency, bodyMass)

			boostedChild := simulation.Rocket{
				PitchDegree:      boostedPitch,
				YawDegree:        boostedYaw,
				Fuel:             boostedFuel,
				BurnRate:         boostedBurnRate,
				EngineEfficiency: boostedEngineEfficiency,
				BodyMass:         bodyMass,
			}
			nextRocketPopulation = append(nextRocketPopulation, boostedChild)

			if len(nextRocketPopulation) == populationSize {
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

		childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency = validationRocket(childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency, bodyMass)

		child := simulation.Rocket{
			PitchDegree:      childPitch,
			YawDegree:        childYaw,
			Fuel:             childFuel,
			BurnRate:         childBurnRate,
			EngineEfficiency: childEngineEfficiency,
			BodyMass:         bodyMass,
		}
		nextRocketPopulation = append(nextRocketPopulation, child)
	}

	return nextRocketPopulation
}
