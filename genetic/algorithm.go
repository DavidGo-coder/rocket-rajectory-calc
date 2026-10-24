package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"time"

	core "github.com/DavidGo-coder/rocket-trajecory-calc"
)

const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
	colorPurple = "\x1b[35m"
	colorCyan   = "\x1b[36m"
	colorBold   = "\x1b[1m"

	portConnection string = "8800"
	hostConnection string = "127.0.0.1"
)

func parseMedians(bufferedReader *bufio.Reader) ([]float64, float64, error) {
	rawMedianHistory, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return nil, 0.0, err
	}

	var medianHistory []float64
	var currentMedian float64

	for i := 0; i < len(rawMedianHistory); i++ {
		median, err := core.ParseFloat(rawMedianHistory, i)
		if err != nil {
			return medianHistory, currentMedian, err
		}

		medianHistory = append(medianHistory, median)

		if len(rawMedianHistory)-1 == i {
			currentMedian = median
		}
	}

	return medianHistory, currentMedian, nil
}

func parseCharaceristics(bufferedReader *bufio.Reader) (int, int, error) {
	characteristics, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return 0, 0, err
	}

	count, err := core.ParseInt(characteristics, 0)
	if err != nil {
		return 0, 0, err
	}

	length, err := core.ParseInt(characteristics, 1)
	if err != nil {
		return 0, 0, err
	}

	return count, length, nil
}

func parseFirstCharacterisics(bufferedReader *bufio.Reader) (int, float64, core.Rocket, error) {
	rawFirstCharacteristics, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return 0, 0.0, core.Rocket{}, err
	}

	populationSize, err := core.ParseInt(rawFirstCharacteristics, 0)
	if err != nil {
		return 0, 0.0, core.Rocket{}, err
	}

	maximumBurnRate, err := core.ParseFloat(rawFirstCharacteristics, 1)
	if err != nil {
		return 0, 0.0, core.Rocket{}, err
	}

	rocket, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return 0, 0.0, core.Rocket{}, err
	}

	firstRocket, err := core.ParseRocketCharacteristics(rocket)
	if err != nil {
		return 0, 0.0, core.Rocket{}, err
	}
	return populationSize, maximumBurnRate, firstRocket, nil
}

func FirstGeneticCalculation(firstRocket core.Rocket, populationSize int, maximumBurnRate float64) []core.Rocket {
	initialPopulation := make([]core.Rocket, 0, populationSize)

	firstRocket = core.ValidationRocket(firstRocket, maximumBurnRate)

	initialPopulation = append(initialPopulation, firstRocket)

	if populationSize > 5 {
		initialPopulation = append(initialPopulation, core.Rocket{PitchDegree: 45.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000.0})

		initialPopulation = append(initialPopulation, core.Rocket{PitchDegree: 35.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: 8.0, EngineEfficiency: 2000.0})

		initialPopulation = append(initialPopulation, core.Rocket{PitchDegree: 65.0, YawDegree: firstRocket.YawDegree, Fuel: 100.0, BurnRate: firstRocket.BurnRate, EngineEfficiency: 2000.0})
	}

	for i := len(initialPopulation); i < populationSize; i++ {
		pitchRand := firstRocket.PitchDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		yawRand := firstRocket.YawDegree + (((rand.Float64() * 2.0) - 1.0) * 5.0)
		fuelRand := firstRocket.Fuel + (((rand.Float64() * 2.0) - 1.0) * (firstRocket.Fuel * 0.1))
		burnRateRand := firstRocket.BurnRate + (((rand.Float64() * 2.0) - 1.0) * 0.5)
		engineEfficiencyRand := firstRocket.EngineEfficiency + ((rand.Float64()*10.0 - 5.0) * 400.0)

		rocketRand := core.Rocket{
			PitchDegree:      pitchRand,
			YawDegree:        yawRand,
			Fuel:             fuelRand,
			BurnRate:         burnRateRand,
			EngineEfficiency: engineEfficiencyRand,
			BodyMass:         firstRocket.BodyMass,
		}

		rocketRand = core.ValidationRocket(rocketRand, maximumBurnRate)
		initialPopulation = append(initialPopulation, rocketRand)
	}
	return initialPopulation
}

func createChild(firstParent, secondParent core.Rocket, bodyMass, maximumBurnRate, medianCoefficient float64, coefficient int) core.Rocket {
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

	child := core.Rocket{
		PitchDegree:      childPitch,
		YawDegree:        childYaw,
		Fuel:             childFuel,
		BurnRate:         childBurnRate,
		EngineEfficiency: childEngineEfficiency,
		BodyMass:         bodyMass,
	}

	child = core.ValidationRocket(child, maximumBurnRate)

	return child
}

func createEmergencyChild(bodyMass, maximumBurnRate float64) core.Rocket {
	ChildPitch := rand.Float64() * 89
	ChildYaw := rand.Float64() * 89
	ChildFuel := rand.Float64() * (bodyMass * 10)
	ChildBurnRate := rand.Float64() * maximumBurnRate
	ChildEngineEfficiency := rand.Float64()*2000 + 500

	child := core.Rocket{
		PitchDegree:      ChildPitch,
		YawDegree:        ChildYaw,
		Fuel:             ChildFuel,
		BurnRate:         ChildBurnRate,
		EngineEfficiency: ChildEngineEfficiency,
		BodyMass:         bodyMass,
	}
	child = core.ValidationRocket(child, maximumBurnRate)
	return child
}

func GeneticCalculation(bestRockets []core.Rocket, populationSize int, medianHistory []float64, currentMedianFitness, bodyMass, maximumBurnRate float64, countGenerations int) []core.Rocket {

	length := len(bestRockets)
	if length == 0 {
		return make([]core.Rocket, populationSize)
	}

	nextRocketPopulation := make([]core.Rocket, length, populationSize)
	copy(nextRocketPopulation, bestRockets)

	sliceTenBestShuffledRocket := make([]core.Rocket, length)
	copy(sliceTenBestShuffledRocket, bestRockets)

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
			nextRocketPopulation = append(nextRocketPopulation, bestRockets[rand.IntN(length)])

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

			continue
		}

		if rand.IntN(100) < 5 {
			mutatedChild := createChild(firstParent, secondParent, bodyMass, maximumBurnRate, 1, 10)
			nextRocketPopulation = append(nextRocketPopulation, mutatedChild)
			if len(nextRocketPopulation) == populationSize {
				break
			}

			continue
		}
		child := createChild(firstParent, secondParent, bodyMass, maximumBurnRate, 1, 1)
		nextRocketPopulation = append(nextRocketPopulation, child)
	}

	return nextRocketPopulation
}

func main() {
	address := net.JoinHostPort(hostConnection, portConnection)

	fmt.Printf("🌐 %s[NETWORK]%s Attempting to reach server at %s...\n", colorCyan, colorReset, address)
	connection, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ %s%s[CONNECTION FAILED]%s Could not connect to remote host: %v\n",
			colorRed, colorBold, colorReset, err)
		return
	}
	defer connection.Close()
	fmt.Printf("✅ %s[CONNECTED]%s Established network connection to server.\n", colorGreen, colorReset)

	reader := bufio.NewReader(connection)
	tools := core.CreateTools(connection, reader)
	defer func() {
		if err := tools.Writer.Flush(); err != nil {
			fmt.Printf("⚠️  %s[IO WARNING]%s Failed to flush data buffer on exit: %v\n", colorYellow, colorReset, err)
		} else {
			fmt.Printf("✨ %s[CLEAN EXIT]%s All pending network data flushed successfully.\n", colorGreen, colorReset)
		}
	}()

	_, err = fmt.Fprintln(tools.Writer, "I|AM|ALGORITHM")
	if err != nil {
		fmt.Printf("❌ %s[HANDSHAKE FAILED]%s Could not send identity packet: %v\n", colorRed, colorReset, err)
		return
	}
	tools.Writer.Flush()
	fmt.Printf("🤝 %s[HANDSHAKE]%s Identification sent successfully: %sI|AM|ALGORITHM%s\n", colorBlue, colorReset, colorPurple, colorReset)

	maximumBurnRate := 0.0
	bodyMass := 0.0
	populationSize := 0

	fmt.Printf("📡 %s[LISTENING]%s Waiting for commands from server...\n", colorCyan, colorReset)
	for {
		readedCommand, err := core.ReadInput(tools.Reader)
		if err != nil {
			fmt.Printf("\n❌ %s[NETWORK ERROR]%s Connection lost with the server: %v\n", colorRed, colorReset, err)
			return
		}

		switch readedCommand {
		case "COMMAND1":
			fmt.Printf("\n⚡ %s[INITIALIZATION]%s Generating starting population...\n", colorBlue, colorReset)

			rawPopulationSize, rawMaximumBurnRate, firstRocket, err := parseFirstCharacterisics(tools.Reader)
			if err != nil {
				fmt.Printf("💥 %s[PARSING ERROR]%s Failed to parse initial characteristics: %v\n", colorRed, colorReset, err)
				continue
			}

			firstRocket = core.ValidationRocket(firstRocket, rawMaximumBurnRate)
			bodyMass = firstRocket.BodyMass
			maximumBurnRate = rawMaximumBurnRate
			populationSize = rawPopulationSize

			population := FirstGeneticCalculation(firstRocket, populationSize, maximumBurnRate)

			for i := 0; i < populationSize; i++ {
				_, err = fmt.Fprintf(tools.Writer, "%v|%v|%v|%v|%v|%v\n",
					population[i].PitchDegree, population[i].YawDegree, population[i].Fuel,
					population[i].BurnRate, population[i].BodyMass, population[i].EngineEfficiency)
				if err != nil {
					fmt.Printf("❌ %s[TRANSMISSION ERROR]%s Failed sending specimen %d: %v\n", colorRed, colorReset, i, err)
					return
				}
				tools.Writer.Flush()
			}
			fmt.Printf("🚀 %s[SUCCESS]%s Initial population of size %s%d%s successfully transmitted.\n",
				colorGreen, colorReset, colorPurple, populationSize, colorReset)

		case "COMMAND2":
			medianHistory, currentMedian, err := parseMedians(tools.Reader)
			if err != nil {
				fmt.Printf("\n💥 %s[PARSING ERROR]%s Failed to parse median history: %v\n", colorRed, colorReset, err)
				continue
			}

			countGenerations, lengthRocket, err := parseCharaceristics(tools.Reader)
			if err != nil {
				fmt.Printf("\n💥 %s[PARSING ERROR]%s Failed to parse generation characteristics: %v\n", colorRed, colorReset, err)
				continue
			}

			bestRockets, err := core.ParseNRocket(lengthRocket, tools.Reader)
			if err != nil {
				fmt.Printf("\n💥 %s[PARSING ERROR]%s Failed parsing best rockets from server: %v\n", colorRed, colorReset, err)
				continue
			}

			if populationSize == 0 || bodyMass == 0 {
				fmt.Printf("\n⚠️  %s[STATE ERROR]%s COMMAND2 received before data initialization. Skipping.\n", colorYellow, colorReset)
				continue
			}

			fmt.Printf("\n🔄 %s[EVOLUTION]%s Processing %sGeneration #%d%s | Best Parent Miss: %s%.2f m%s\n",
				colorCyan, colorReset, colorPurple, countGenerations, colorReset, colorYellow, currentMedian, colorReset)

			population := GeneticCalculation(bestRockets, populationSize, medianHistory, currentMedian, bodyMass, maximumBurnRate, countGenerations)

			for i := 0; i < populationSize; i++ {
				_, err = fmt.Fprintf(tools.Writer, "%v|%v|%v|%v|%v|%v\n",
					population[i].PitchDegree, population[i].YawDegree, population[i].Fuel,
					population[i].BurnRate, population[i].BodyMass, population[i].EngineEfficiency)
				if err != nil {
					fmt.Printf("❌ %s[TRANSMISSION ERROR]%s Failed sending generation specimen %d: %v\n", colorRed, colorReset, i, err)
					return
				}
				tools.Writer.Flush()
			}
			fmt.Printf("🧬 %s[SUCCESS]%s Generation %s%d%s evolved. New DNA payload sent back.\n",
				colorGreen, colorReset, colorPurple, countGenerations, colorReset)

		case "FINISHWORKING":
			fmt.Printf("\n🛑 %s[SHUTDOWN]%s Received termination command from server. Exiting gracefully...\n", colorYellow, colorReset)
			return

		default:
			fmt.Printf("\n⚠️  %s[UNKNOWN COMMAND]%s Received unexpected instruction from server: %q\n", colorYellow, colorReset, readedCommand)
		}
	}
}
