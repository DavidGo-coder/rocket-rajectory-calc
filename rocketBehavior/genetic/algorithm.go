package genetic

import (
	"math"
	"math/rand/v2"

	"github.com/DavidGo-coder/rocket-trajectory-calc/simulation"
)

func ValidationRocket(pitchDegree, yawDegree, fuel, burnRate, engineEfficiency, bodyMass, maximumBurnRate float64) (float64, float64, float64, float64, float64) {
	if pitchDegree <= 1 {
		pitchDegree = 1
	}
	if pitchDegree >= 89 {
		pitchDegree = 89
	}
	if yawDegree <= 1 {
		yawDegree = 1
	}
	if yawDegree >= 89 {
		yawDegree = 89
	}
	if fuel < 5.0 {
		fuel = 5.0
	}
	if fuel > bodyMass*10.0 {
		fuel = bodyMass * 10.0
	}
	if engineEfficiency < 500.0 {
		engineEfficiency = 500.0
	}
	if engineEfficiency > 2500.0 {
		engineEfficiency = 2500.0
	}
	if engineEfficiency*burnRate < (bodyMass+fuel)*simulation.GravitationalConstant {
		burnRate = (bodyMass + fuel) * simulation.GravitationalConstant / engineEfficiency
	}
	if burnRate > maximumBurnRate {
		burnRate = maximumBurnRate
	}
	return pitchDegree, yawDegree, fuel, burnRate, engineEfficiency
}

func FirstGeneticCalculation(firstRocket simulation.Rocket, populationSize int, maximumBurnRate float64) []simulation.Rocket {
	initialPopulation := make([]simulation.Rocket, 0, populationSize)

	firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency = ValidationRocket(firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.EngineEfficiency, firstRocket.BodyMass, maximumBurnRate)

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

		pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand = ValidationRocket(pitchRand, yawRand, fuelRand, burnRateRand, engineEfficiencyRand, firstRocket.BodyMass, maximumBurnRate)

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

func createChild(firstParent, secondParent simulation.Rocket, bodyMass, maximumBurnRate, medianCoefficient float64, coefficient int) simulation.Rocket {
	childPitchRand := rand.Float64()
	childYawRand := rand.Float64()
	childFuelRand := rand.Float64()
	childBurnRateRand := rand.Float64()
	childEngineEfficiencyRand := rand.Float64()

	childPitch := (firstParent.PitchDegree*childPitchRand + secondParent.PitchDegree*(1.0-childPitchRand)) + rand.Float64()*6.0*float64(coefficient) - 3.0*float64(coefficient)
	childYaw := (firstParent.YawDegree*childYawRand + secondParent.YawDegree*(1.0-childYawRand)) + rand.Float64()*6.0*float64(coefficient) - 3.0*float64(coefficient)
	childFuel := (firstParent.Fuel*childFuelRand+secondParent.Fuel*(1.0-childFuelRand))*medianCoefficient + rand.Float64()*4.0*float64(coefficient) - 2.0*float64(coefficient)
	childBurnRate := math.Max(0.5, (firstParent.BurnRate*childBurnRateRand+secondParent.BurnRate*(1.0-childBurnRateRand))*medianCoefficient+rand.Float64()*0.4*float64(coefficient)-0.2*float64(coefficient))
	childEngineEfficiency := (firstParent.EngineEfficiency*childEngineEfficiencyRand+secondParent.EngineEfficiency*(1.0-childEngineEfficiencyRand))*medianCoefficient + rand.Float64()*10.0*float64(coefficient) - 5.0*float64(coefficient)

	childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency = ValidationRocket(childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency, bodyMass, maximumBurnRate)

	child := simulation.Rocket{
		PitchDegree:      childPitch,
		YawDegree:        childYaw,
		Fuel:             childFuel,
		BurnRate:         childBurnRate,
		EngineEfficiency: childEngineEfficiency,
		BodyMass:         bodyMass,
	}
	return child
}

func createEmergencyChild(bodyMass, maximumBurnRate float64) simulation.Rocket {
	ChildPitch := rand.Float64() * 89
	ChildYaw := rand.Float64() * 89
	ChildFuel := rand.Float64() * (bodyMass * 10)
	ChildBurnRate := rand.Float64() * maximumBurnRate
	ChildEngineEfficiency := rand.Float64()*2000 + 500
	ChildPitch, ChildYaw, ChildFuel, ChildBurnRate, ChildEngineEfficiency = ValidationRocket(ChildPitch, ChildYaw, ChildFuel, ChildBurnRate, ChildEngineEfficiency, bodyMass, maximumBurnRate)

	child := simulation.Rocket{
		PitchDegree:      ChildPitch,
		YawDegree:        ChildYaw,
		Fuel:             ChildFuel,
		BurnRate:         ChildBurnRate,
		EngineEfficiency: ChildEngineEfficiency,
		BodyMass:         bodyMass,
	}
	return child
}

func GeneticCalculation(sliceTenBestRocket []simulation.Rocket, populationSize int, medianHistory []float64, currentMedianFitness, bodyMass, maximumBurnRate float64, countGenerations int) []simulation.Rocket {
	//nextRocketPopulation := make([]simulation.Rocket, 0, populationSize)

	length := len(sliceTenBestRocket)
	if length == 0 {
		return make([]simulation.Rocket, populationSize)
	}

	nextRocketPopulation := make([]simulation.Rocket, length, populationSize)
	copy(nextRocketPopulation, sliceTenBestRocket)

	//for i := 0; i < length && len(nextRocketPopulation) < populationSize; i++ {
	//	nextRocketPopulation = append(nextRocketPopulation, sliceTenBestRocket[i])
	//}

	sliceTenBestShuffledRocket := make([]simulation.Rocket, length)
	copy(sliceTenBestShuffledRocket, sliceTenBestRocket)

	rand.Shuffle(length, func(i, j int) {
		sliceTenBestShuffledRocket[i], sliceTenBestShuffledRocket[j] = sliceTenBestShuffledRocket[j], sliceTenBestShuffledRocket[i]
	})

	medianCoefficient := currentMedianFitness/500 + 1.0

	if countGenerations >= 4 {
		minimumMedian, maximumMedian := medianHistory[0], medianHistory[0]
		for _, m := range medianHistory {
			if minimumMedian > m {
				minimumMedian = m
			}
			if maximumMedian < m {
				maximumMedian = m
			}
		}

		if maximumMedian-minimumMedian < currentMedianFitness*0.001 {
			nextRocketPopulation = append(nextRocketPopulation, sliceTenBestRocket[rand.IntN(length)])

			for len(nextRocketPopulation) < populationSize {
				emergencyChild := createEmergencyChild(bodyMass, maximumBurnRate)

				nextRocketPopulation = append(nextRocketPopulation, emergencyChild)
			}
			return nextRocketPopulation
		}
	}

	parentIndex := 0
	for len(nextRocketPopulation) < populationSize {
		firstParent := sliceTenBestShuffledRocket[parentIndex%length]
		secondParent := sliceTenBestShuffledRocket[(parentIndex+1)%length]
		parentIndex++

		if len(nextRocketPopulation)%5 == 0 {
			boostedChild := createChild(firstParent, secondParent, bodyMass, maximumBurnRate, medianCoefficient, 1)
			nextRocketPopulation = append(nextRocketPopulation, boostedChild)

			if len(nextRocketPopulation) == populationSize {
				break
			}
		}

		if rand.IntN(100) < 5 {
			mutatedChild := createChild(firstParent, secondParent, bodyMass, maximumBurnRate, 1, 10)
			nextRocketPopulation = append(nextRocketPopulation, mutatedChild)

			continue
		}
		child := createChild(firstParent, secondParent, bodyMass, maximumBurnRate, 1, 1)
		nextRocketPopulation = append(nextRocketPopulation, child)
	}

	return nextRocketPopulation
}
