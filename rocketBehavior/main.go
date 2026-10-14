package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DavidGo-coder/rocket-trajectory-calc/genetic"
	"github.com/DavidGo-coder/rocket-trajectory-calc/simulation"
)

const (
	portConnection    string = "8800"
	timeOneGeneration        = 3 * time.Second
	medianHisorySize  int    = 5
)

func readInput(bufferedReader *bufio.Reader) ([]string, error) {
	readedData, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	splitedData := strings.Split(readedData, ",")
	if err == io.EOF {
		err = nil
	}
	return splitedData, err
}

func parseInt(slice []string, position int) (int, error) {
	result, err := strconv.ParseInt(strings.TrimSpace(slice[position]), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(result), err
}

func parseFloat(slice []string, position int) (float64, error) {
	result, err := strconv.ParseFloat(strings.TrimSpace(slice[position]), 64)
	if err != nil {
		return 0.0, err
	}
	return result, err
}

func verticalBar(start, symbol, finish string) {
	fmt.Println(start + strings.Repeat(symbol, 50) + finish)
}

func readTarget(bufferedReader *bufio.Reader) (targetX, targetY, targetZ int, err error) {
	for {
		verticalBar("\n┌", "─", "┐")
		fmt.Println("│ 🎯 TARGET COORDINATES SPECIFICATION              │")
		verticalBar("└", "─", "┘")
		fmt.Println(" Enter position in meters using format: X, Y, Z")
		fmt.Print(" ❯ ")

		threeTarget, err := readInput(bufferedReader)
		if err != nil {
			log.Printf("⚠️ Critical IO Error: %v", err)
			return targetX, targetY, targetZ, err
		}

		if len(threeTarget) != 3 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 3 coordinates (X,Y,Z).")
			verticalBar("", "─", "")
			continue
		}

		rawX, err := parseInt(threeTarget, 0)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse X: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawY, err := parseInt(threeTarget, 1)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Y: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawZ, err := parseInt(threeTarget, 2)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Z: %v", err)
			verticalBar("", "─", "")
			continue
		}

		if (rawX > 0 && rawY >= 0 && rawZ >= 0) || (rawX >= 0 && rawY >= 0 && rawZ > 0) {
			fmt.Println("✨ [SUCCESS] Target coordinates successfully validated.")
			verticalBar("", "─", "")
			targetX = rawX
			targetY = rawY
			targetZ = rawZ
			break
		} else {
			fmt.Printf("⚠️  [VALIDATION FAILED] At least X or Z must be positive.\n"+
				"    Y cannot be the only active axis.\n"+
				"    Got: X=%v, Y=%v, Z=%v\n", rawX, rawY, rawZ)
			verticalBar("", "─", "")
			continue
		}
	}

	return targetX, targetY, targetZ, err
}

func readFirstRocket(bufferedReader *bufio.Reader, x, y int) (rocket simulation.Rocket, err error) {

	pitch, yaw, fuel, burnRate, bodyMass := 0.0, 0.0, 0.0, 0.0, 0.0

	for {
		verticalBar("\n┌", "─", "┐")
		fmt.Println("│ 🚀 INITIAL ROCKET CONFIGURATION (CORE DNA)       │")
		verticalBar("└", "─", "┘")
		fmt.Println(" Enter 5 metrics separated by commas:")
		fmt.Println(" Pitch(deg), Yaw(deg), Fuel(kg), BurnRate(kg/s), BodyMass(kg)")
		fmt.Print(" ❯ ")

		bothParameters, err := readInput(bufferedReader)
		if err != nil {
			log.Printf("⚠️ Critical IO Error: %v", err)
			return rocket, err
		}

		if len(bothParameters) != 5 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 5 parameters.")
			verticalBar("", "─", "")
			continue
		}

		rawPitchFirstRocket, err := parseFloat(bothParameters, 0)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Pitch: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawYawFirstRocket, err := parseFloat(bothParameters, 1)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Yaw: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawFuelFirstRocket, err := parseFloat(bothParameters, 2)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Fuel mass: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawBurnRateFirstRocket, err := parseFloat(bothParameters, 3)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Burn rate: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawBodyMassFirstRocket, err := parseFloat(bothParameters, 4)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Body mass: %v", err)
			verticalBar("", "─", "")
			continue
		}

		if rawYawFirstRocket >= 1 && rawYawFirstRocket <= 89 && rawPitchFirstRocket >= 1 && rawPitchFirstRocket <= 89 && rawFuelFirstRocket > 0 && rawBurnRateFirstRocket > 0 && rawBodyMassFirstRocket > 0 {
			fmt.Println("✨ [SUCCESS] Rocket parameters successfully validated.")
			verticalBar("", "─", "")
			pitch = rawPitchFirstRocket
			yaw = rawYawFirstRocket
			fuel = rawFuelFirstRocket
			burnRate = rawBurnRateFirstRocket
			bodyMass = rawBodyMassFirstRocket
			break
		} else {
			fmt.Printf("⚠️ [VALIDATION FAILED] Out of range! Angles (%v°, %v°) must be 1-89°, other metrics (%v, %v, %v) must be positive.\n", rawPitchFirstRocket, rawYawFirstRocket, rawBodyMassFirstRocket, rawBurnRateFirstRocket, rawFuelFirstRocket)
			verticalBar("", "─", "")
			continue
		}
	}

	calculatedEngineEfficiency := 800.0 + (burnRate/bodyMass)*1500.0

	distance := math.Hypot(float64(x), float64(y))
	maximumBurnRate := 20.0 + (distance / 50.0)

	pitch, yaw, fuel, burnRate, calculatedEngineEfficiency = genetic.ValidationRocket(pitch, yaw, fuel, burnRate, calculatedEngineEfficiency, bodyMass, maximumBurnRate)

	rocket = simulation.Rocket{PitchDegree: pitch, YawDegree: yaw, Fuel: fuel, BurnRate: burnRate, EngineEfficiency: calculatedEngineEfficiency, BodyMass: bodyMass}

	return rocket, err
}

func readPopulationParameters(bufferedReader *bufio.Reader) (generations, populationSize int, err error) {
	for {
		verticalBar("\n┌", "─", "┐")
		fmt.Println("│ 👥 EVOLUTIONARY POPULATION CONSTRAINTS           │")
		verticalBar("└", "─", "┘")
		fmt.Println(" Enter parameters separated by commas:")
		fmt.Println(" PopulationSize (min: 2), TotalGenerations")
		fmt.Print(" ❯ ")

		bothConditions, err := readInput(bufferedReader)
		if err != nil {
			log.Printf("⚠️ Critical IO Error: %v", err)
			return generations, populationSize, err
		}

		if len(bothConditions) != 2 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 2 parameters.")
			verticalBar("", "─", "")
			continue
		}

		rawPopSize, err := parseInt(bothConditions, 0)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Population size: %v", err)
			verticalBar("", "─", "")
			continue
		}

		rawGenerations, err := parseInt(bothConditions, 1)
		if err != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Total generations: %v", err)
			verticalBar("", "─", "")
			continue
		}

		if rawPopSize > 1 && rawGenerations > 0 {
			fmt.Println("✨ [SUCCESS] Evolution parameters successfully validated.")
			verticalBar("", "─", "")
			populationSize = rawPopSize
			generations = rawGenerations
			break
		} else {
			fmt.Printf("⚠️ [VALIDATION FAILED] Invalid constraints! Size (%v) must be > 1, Generations (%v) must be > 0.\n", rawPopSize, rawGenerations)
			verticalBar("", "─", "")
			continue
		}
	}
	return generations, populationSize, err
}

func ReadAllInput() (targetX, targetY, targetZ int, rocket simulation.Rocket, populationSize, generation int, err error) {
	bufferedReader := bufio.NewReader(os.Stdin)

	targetX, targetY, targetZ, err = readTarget(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generation, err
	}

	firstRocket, err := readFirstRocket(bufferedReader, targetX, targetY)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generation, err
	}

	generations, populationSize, err := readPopulationParameters(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generation, err
	}

	verticalBar("", "=", "")
	fmt.Println("🚀 INITIALIZING GENETIC SIMULATION ENGINE... 🚀")
	verticalBar("", "=", "")
	fmt.Printf("🎯 Target Lock:     X=%v, Y=%v, Z=%v\n", targetX, targetY, targetZ)
	fmt.Printf("👥 Population:    %v specimens | %v generations\n", populationSize, generations)
	fmt.Printf("🧬 First DNA: Pitch=%.2f°, Yaw=%.2f°, Fuel=%.2f kg, Mass=%.2f kg, Burn=%.2f kg/s, Eff=%.2f\n", firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BodyMass, firstRocket.BurnRate, firstRocket.EngineEfficiency)
	verticalBar("", "=", "")

	return targetX, targetY, targetZ, firstRocket, populationSize, generations, nil
}

func WaitSlowPython(connectionReader *bufio.Reader) error {
	currentTime := time.Now()

	_, err := connectionReader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("❌ [SYNC ERROR] Python client disconnected or failed to respond!\n")
		return err
	}

	if time.Since(currentTime) < timeOneGeneration {
		timeSpent := time.Since(currentTime)
		time.Sleep(timeOneGeneration - timeSpent)
	}
	fmt.Printf("✨ [SYNC] Python pipeline processed generation data successfully!")
	return err
}

func AcceptPython() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+portConnection)
	if err != nil {
		return listener, err
	}

	fmt.Printf("📡 [NET SYSTEM] TCP Server started successfully on port %v\n", portConnection)

	return listener, err
}

func main() {
	targetX, targetY, targetZ, firstRocket, populationSize, generations, err := ReadAllInput()
	if err != nil {
		return
	}

	listener, err := AcceptPython()
	if err != nil {
		log.Printf("Error creating connection with Python, %v\n", err)
		return
	}

	defer listener.Close()

	connection, err := listener.Accept()
	if err != nil {
		log.Printf("❌ Python client handshake failed: %v", err)
		return
	}

	defer connection.Close()
	connectionReader := bufio.NewReader(connection)
	fmt.Printf("🌐 [NET ENGINE] Spawning TCP Server on port %v...\n", portConnection)

	_, err = fmt.Fprintf(connection, "TARGET|%v,%v\n", targetX, targetY)
	if err != nil {
		log.Println("📡❌ [NET ERROR] Failed to transmit target coordinates packet.")
		return
	}

	//targetDistance := math.Sqrt(math.Pow(float64(targetCoordinateX), 2) + math.Pow(float64(targetCoordinateY), 2) + math.Pow(float64(targetCoordinateZ), 2))
	targetDistance := math.Hypot(float64(targetX), float64(targetY))
	maximumBurnRate := 20.0 + (targetDistance / 50.0)

	rocketPopulation := genetic.FirstGeneticCalculation(firstRocket, populationSize, maximumBurnRate)

	sliceTenBestRocket, medianPopulation := simulation.CalculationTrajectory(targetX, targetY, targetZ, populationSize, rocketPopulation)
	initialMedian := medianPopulation

	for id, rocket := range sliceTenBestRocket {
		if rocket.TrajectoryCSV == "" {
			rocket.TrajectoryCSV = "0.00,0.00"
		}

		_, err := fmt.Fprintf(connection, "ROCKET_ID:%v;%v\n", id+1, rocket.TrajectoryCSV)
		if err != nil {
			log.Println("📡⚠️ [NET WARNING] Payload transmission interrupted.")
			return
		}
	}

	err = WaitSlowPython(connectionReader)
	if err != nil {
		fmt.Printf("❌ [SYNC ERROR] Python client disconnected or failed to respond!\n")
		return
	}

	medianHistory := make([]float64, medianHisorySize)
	medianHistory[0] = initialMedian

	rocketPopulation2 := genetic.GeneticCalculation(sliceTenBestRocket, populationSize, medianHistory, initialMedian, firstRocket.BodyMass, maximumBurnRate, 0)

	adaptiveThreshold := math.Max(80.0, initialMedian*0.005)

	for countGenerations := 0; countGenerations < generations; countGenerations++ {
		fmt.Printf("⏳ [EVOLUTION] Simulation of generation %v/%v running...\n", countGenerations+1, generations)

		sliceTenBestRocket2, medianNextFitness := simulation.CalculationTrajectory(targetX, targetY, targetZ, populationSize, rocketPopulation2)

		medianHistory[(countGenerations+1)%medianHisorySize] = medianNextFitness

		for id, rocket := range sliceTenBestRocket2 {
			if rocket.TrajectoryCSV == "" {
				rocket.TrajectoryCSV = "0.00,0.00"
			}

			_, err = fmt.Fprintf(connection, "ROCKET_ID:%v;%v\n", id+1, rocket.TrajectoryCSV)
			if err != nil {
				fmt.Println("📡💥 [NET CRITICAL] Core transmission failure. Aborting.")
				return
			}

			if rocket.Fitness <= adaptiveThreshold && rocket.Y > 5 {
				fmt.Println("\n==================================================")
				fmt.Println("🎯 THE GOAL WAS ACHIEVED! WE HAVE A WINNER! 🎯")
				fmt.Println("==================================================")
				fmt.Printf("🚀 Best Rocket Characteristics:\n")
				fmt.Printf("   • Pitch Angle:   %.2f degrees\n", rocket.PitchDegree)
				fmt.Printf("   • Initial Fuel:  %.2f kg\n", rocket.Fuel)
				fmt.Printf("   • Engine Burn:   %.2f kg/s\n", rocket.BurnRate)
				fmt.Printf("--------------------------------------------------\n")
				fmt.Printf("📊 Flight Performance:\n")
				fmt.Printf("   • Final X Pos:   %.2f meters\n", rocket.X)
				fmt.Printf("   • Max Y Pos:  %.2f meters (Max Y)\n", rocket.Y)
				fmt.Printf("   • Target Miss:   %.2f meters\n", rocket.Fitness)
				fmt.Println("==================================================")

				fmt.Println("\n👋 [SYSTEM] Simulation successfully completed. Closing connections...")
				time.Sleep(1 * time.Second)
				return
			}
		}

		err = WaitSlowPython(connectionReader)
		if err != nil {
			fmt.Printf("❌ [SYNC ERROR] Python client disconnected or failed to respond!\n")
			return
		}
		rocketPopulation2 = genetic.GeneticCalculation(sliceTenBestRocket2, populationSize, medianHistory, medianNextFitness, firstRocket.BodyMass, maximumBurnRate, countGenerations+1)
	}
}
