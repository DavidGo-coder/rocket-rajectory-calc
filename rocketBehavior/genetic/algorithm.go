package genetic

import (
	"math"
	"math/rand/v2"

	"github.com/DavidGo-coder/rocket-trajectory-calc/simulation"
)

func ValidationRocket(pitchDegree, yawDegree, fuel, burnRate, engineEfficiency, bodyMass, maximumBurnRate float64) (float64, float64, float64, float64, float64) {
	if pitchDegree < 1 {
		pitchDegree = 1
	}
	if pitchDegree > 89 {
		pitchDegree = 89
	}
	if yawDegree < 1 {
		yawDegree = 1
	}
	if yawDegree > 89 {
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

func GeneticCalculation(sliceTenBestRocket []simulation.Rocket, populationSize int, medianHistory []float64, currentMedianFitness, bodyMass, maximumBurnRate float64, countGenerations int) []simulation.Rocket {
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
			nextRocketPopulation = append(nextRocketPopulation, sliceTenBestRocket[int(rand.Float64()*10)])

			for len(nextRocketPopulation) < populationSize {

				EmergencyChildPitch := rand.Float64() * 89
				EmergencyChildYaw := rand.Float64() * 89
				EmergencyChildFuel := rand.Float64() * (bodyMass * 10)
				EmergencyChildBurnRate := rand.Float64() * maximumBurnRate
				EmergencyChildEngineEfficiency := rand.Float64()*2000 + 500
				EmergencyChildPitch, EmergencyChildYaw, EmergencyChildFuel, EmergencyChildBurnRate, EmergencyChildEngineEfficiency = ValidationRocket(EmergencyChildPitch, EmergencyChildYaw, EmergencyChildFuel, EmergencyChildBurnRate, EmergencyChildEngineEfficiency, bodyMass, maximumBurnRate)

				child := simulation.Rocket{
					PitchDegree:      EmergencyChildPitch,
					YawDegree:        EmergencyChildYaw,
					Fuel:             EmergencyChildFuel,
					BurnRate:         EmergencyChildBurnRate,
					EngineEfficiency: EmergencyChildEngineEfficiency,
					BodyMass:         bodyMass,
				}
				nextRocketPopulation = append(nextRocketPopulation, child)
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
			boostedPitchRand := rand.Float64()
			boostedYawRand := rand.Float64()
			boostedFuelRand := rand.Float64()
			boostedBurnRateRand := rand.Float64()
			boostedEngineEfficiencyRand := rand.Float64()

			boostedPitch := (firstParent.PitchDegree*boostedPitchRand + secondParent.PitchDegree*(1.0-boostedPitchRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			boostedYaw := (firstParent.YawDegree*boostedYawRand + secondParent.YawDegree*(1.0-boostedYawRand)) + (((rand.Float64() * 2.0) - 1.0) * 5.0)
			boostedFuel := (firstParent.Fuel*boostedFuelRand+secondParent.Fuel*(1.0-boostedFuelRand))*(currentMedianFitness/500+1.0) + rand.Float64()*4.0 - 2.0
			boostedBurnRate := math.Max(0.5, (firstParent.BurnRate*boostedBurnRateRand+secondParent.BurnRate*(1.0-boostedBurnRateRand))*(currentMedianFitness/500+1.0)+rand.Float64()*0.4-0.2)
			boostedEngineEfficiency := (firstParent.EngineEfficiency*boostedEngineEfficiencyRand+secondParent.EngineEfficiency*(1.0-boostedEngineEfficiencyRand))*(currentMedianFitness/500+1.0) + rand.Float64()*10.0 - 5.0

			boostedPitch, boostedYaw, boostedFuel, boostedBurnRate, boostedEngineEfficiency = ValidationRocket(boostedPitch, boostedYaw, boostedFuel, boostedBurnRate, boostedEngineEfficiency, bodyMass, maximumBurnRate)

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

		if rand.IntN(100) < 5 {
			mutatedChildPitchRand := rand.Float64()
			mutatedChildYawRand := rand.Float64()
			mutatedChildFuelRand := rand.Float64()
			mutatedChildBurnRateRand := rand.Float64()
			mutatedChildEngineEfficiencyRand := rand.Float64()

			mutatedChildPitch := (firstParent.PitchDegree*mutatedChildPitchRand + secondParent.PitchDegree*(1.0-mutatedChildPitchRand)) + rand.Float64()*60.0 - 30.0
			mutatedChildYaw := (firstParent.YawDegree*mutatedChildYawRand + secondParent.YawDegree*(1.0-mutatedChildYawRand)) + rand.Float64()*60.0 - 30.0
			mutatedChildFuel := (firstParent.Fuel*mutatedChildFuelRand + secondParent.Fuel*(1.0-mutatedChildFuelRand)) + rand.Float64()*40.0 - 20.0
			mutatedChildBurnRate := math.Max(0.5, (firstParent.BurnRate*mutatedChildBurnRateRand+secondParent.BurnRate*(1.0-mutatedChildBurnRateRand))+rand.Float64()*4-2)
			mutatedChildEngineEfficiency := (firstParent.EngineEfficiency*mutatedChildEngineEfficiencyRand + secondParent.EngineEfficiency*(1.0-mutatedChildEngineEfficiencyRand)) + rand.Float64()*100.0 - 50.0

			mutatedChildPitch, mutatedChildYaw, mutatedChildFuel, mutatedChildBurnRate, mutatedChildEngineEfficiency = ValidationRocket(mutatedChildPitch, mutatedChildYaw, mutatedChildFuel, mutatedChildBurnRate, mutatedChildEngineEfficiency, bodyMass, mutatedChildBurnRate)

			mutatedChild := simulation.Rocket{
				PitchDegree:      mutatedChildPitch,
				YawDegree:        mutatedChildYaw,
				Fuel:             mutatedChildFuel,
				BurnRate:         mutatedChildBurnRate,
				EngineEfficiency: mutatedChildEngineEfficiency,
				BodyMass:         bodyMass,
			}
			nextRocketPopulation = append(nextRocketPopulation, mutatedChild)

			continue
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

		childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency = ValidationRocket(childPitch, childYaw, childFuel, childBurnRate, childEngineEfficiency, bodyMass, maximumBurnRate)

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
