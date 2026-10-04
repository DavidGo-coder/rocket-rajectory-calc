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

func LaunchPopulation(firstRocket simulation.Rocket, targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize, generations int) {
	listener, errorListening := net.Listen("tcp", ":8080")
	if errorListening != nil {
		log.Fatalf("❌ Failed to bind TCP port 8080: %v\n", errorListening)
	}
	defer listener.Close()

	fmt.Println("📡 [NET SYSTEM] TCP Server started successfully on port 8080")

	connection, errorConection := listener.Accept()
	if errorConection != nil {
		log.Printf("❌ Python client handshake failed: %v", errorConection)
		return
	}
	defer connection.Close()

	fmt.Println("✨ [CONNECTION] Python neural-radar pipeline connected!")

	bothTargetPacket := fmt.Sprintf("TARGET|%d,%d\n", targetCoordinateX, targetCoordinateY)
	connection.Write([]byte(bothTargetPacket))
	time.Sleep(150 * time.Millisecond)

	rocketPopulation := genetic.FirstGeneticCalculation(firstRocket, populationSize)

	sliceTenBestRocket, medianPopulation := simulation.CalculationTrajectory(targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize, rocketPopulation)
	initialMedian := medianPopulation

	for id, rocket := range sliceTenBestRocket {
		dataCSV := fmt.Sprintf("ROCKET_ID:%d;%s\n", id+1, rocket.TrajectoryCSV)
		connection.Write([]byte(dataCSV))
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(3000 * time.Millisecond)

	rocketPopulation2 := genetic.GeneticCalculation(sliceTenBestRocket, populationSize, medianPopulation, firstRocket.BodyMass)

	for countGenerations := 0; countGenerations < generations; countGenerations++ {
		fmt.Printf("⏳ [EVOLUTION] Simulation of generation %d/%d running...\n", countGenerations+1, generations)
		sliceTenBestRocket2, medianNextFitness := simulation.CalculationTrajectory(targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize, rocketPopulation2)

		adaptiveThreshold := math.Max(80.0, initialMedian*0.005)

		for id, rocket := range sliceTenBestRocket2 {
			if rocket.TrajectoryCSV == "" {
				rocket.TrajectoryCSV = "0.00,0.00"
			}

			dataCSV := fmt.Sprintf("ROCKET_ID:%d;%s\n", id+1, rocket.TrajectoryCSV)

			_, errorWriting2 := connection.Write([]byte(dataCSV))
			if errorWriting2 != nil && errorWriting2 != io.EOF {
				return
			}

			if rocket.Fitness <= adaptiveThreshold && rocket.X > 10 {
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
			time.Sleep(20 * time.Millisecond)
		}

		time.Sleep(3000 * time.Millisecond)
		rocketPopulation2 = genetic.GeneticCalculation(sliceTenBestRocket2, populationSize, medianNextFitness, firstRocket.BodyMass)
	}
}

func main() {
	bufferedReader := bufio.NewReader(os.Stdin)

	targetCoordinateX := 0
	targetCoordinateY := 0
	targetCoordinateZ := 0
	pitchFirstRocket := 0.0
	yawFirstRocket := 0.0
	fuelFirstRocket := 0.0
	burnRateFirstRocket := 0.0
	bodyMassPopulation := 0.0
	populationSize := 0
	generations := 0

	for {
		fmt.Println("┌──────────────────────────────────────────────────┐")
		fmt.Println("│ 🎯 TARGET COORDINATES SPECIFICATION              │")
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println(" Enter position in meters using format: X, Y, Z")
		fmt.Print(" ❯ ")

		bothTargetCoordinates, errorReading1 := bufferedReader.ReadString('\n')
		if errorReading1 != nil && errorReading1 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading1)
			return
		}
		bothTargetCoordinates = strings.TrimSpace(bothTargetCoordinates)
		targetSlice := strings.Split(bothTargetCoordinates, ",")

		if len(targetSlice) != 3 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 3 coordinates (X,Y,Z).")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawTargetCoordinateX, errorTransforming1 := strconv.ParseInt(strings.TrimSpace(targetSlice[0]), 10, 64)
		if errorTransforming1 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse X: %v", errorTransforming1)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawTargetCoordinateY, errorTransforming2 := strconv.ParseInt(strings.TrimSpace(targetSlice[1]), 10, 64)
		if errorTransforming2 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Y: %v", errorTransforming2)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawTargetCoordinateZ, errorTransforming3 := strconv.ParseInt(strings.TrimSpace(targetSlice[2]), 10, 64)
		if errorTransforming3 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Z: %v", errorTransforming3)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if (rawTargetCoordinateX > 0 && rawTargetCoordinateY >= 0 && rawTargetCoordinateZ >= 0) || (rawTargetCoordinateX >= 0 && rawTargetCoordinateY >= 0 && rawTargetCoordinateZ > 0) {
			fmt.Println("✨ [SUCCESS] Target coordinates successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			targetCoordinateX = int(rawTargetCoordinateX)
			targetCoordinateY = int(rawTargetCoordinateY)
			targetCoordinateZ = int(rawTargetCoordinateZ)
			break
		} else {
			fmt.Printf("⚠️  [VALIDATION FAILED] Non-zero coordinate required. Got: X=%v, Y=%v, Z=%v\n", targetCoordinateX, targetCoordinateY, targetCoordinateZ)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}
	}

	for {
		fmt.Println("\n┌──────────────────────────────────────────────────┐")
		fmt.Println("│ 🚀 INITIAL ROCKET CONFIGURATION (CORE DNA)       │")
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println(" Enter 5 metrics separated by commas:")
		fmt.Println(" Pitch(deg), Yaw(deg), Fuel(kg), BurnRate(kg/s), BodyMass(kg)")
		fmt.Print(" ❯ ")

		bothConditionsFirstRocket, errorReading2 := bufferedReader.ReadString('\n')
		if errorReading2 != nil && errorReading2 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading2)
			return
		}
		bothConditionsFirstRocket = strings.TrimSpace(bothConditionsFirstRocket)
		sliceConditionsFirstRocket := strings.Split(bothConditionsFirstRocket, ",")

		if len(sliceConditionsFirstRocket) != 5 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 5 parameters.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawPitchFirstRocket, errorTransforming4 := strconv.ParseFloat(strings.TrimSpace(sliceConditionsFirstRocket[0]), 64)
		if errorTransforming4 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Pitch: %v", errorTransforming4)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawYawFirstRocket, errorTransforming5 := strconv.ParseFloat(strings.TrimSpace(sliceConditionsFirstRocket[1]), 64)
		if errorTransforming5 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Yaw: %v", errorTransforming5)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawFuelFirstRocket, errorTransforming6 := strconv.ParseFloat(strings.TrimSpace(sliceConditionsFirstRocket[2]), 64)
		if errorTransforming6 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Fuel mass: %v", errorTransforming6)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawBurnRateFirstRocket, errorTransforming7 := strconv.ParseFloat(strings.TrimSpace(sliceConditionsFirstRocket[3]), 64)
		if errorTransforming7 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Burn rate: %v", errorTransforming7)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawBodyMassPopulation, errorTransforming8 := strconv.ParseFloat(strings.TrimSpace(sliceConditionsFirstRocket[4]), 64)
		if errorTransforming8 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Body mass: %v", errorTransforming8)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if rawYawFirstRocket > 1 && rawYawFirstRocket < 89 && rawPitchFirstRocket > 1 && rawPitchFirstRocket < 89 && rawFuelFirstRocket > 0 && rawBurnRateFirstRocket > 0 && rawBodyMassPopulation > 0 {
			fmt.Println("✨ [SUCCESS] Rocket parameters successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			pitchFirstRocket = rawPitchFirstRocket
			yawFirstRocket = rawYawFirstRocket
			fuelFirstRocket = rawFuelFirstRocket
			burnRateFirstRocket = rawBurnRateFirstRocket
			bodyMassPopulation = rawBodyMassPopulation
			break
		} else {
			fmt.Printf("⚠️ [VALIDATION FAILED] Out of range! Angles (%v°, %v°) must be 1-89°, other metrics (%v, %v, %v) must be positive.\n", pitchFirstRocket, yawFirstRocket, bodyMassPopulation, burnRateFirstRocket, fuelFirstRocket)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}
	}

	for {
		fmt.Println("\n┌──────────────────────────────────────────────────┐")
		fmt.Println("│ 👥 EVOLUTIONARY POPULATION CONSTRAINTS           │")
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println(" Enter parameters separated by commas:")
		fmt.Println(" PopulationSize (min: 2), TotalGenerations")
		fmt.Print(" ❯ ")

		bothConditionsPopulation, errorReading3 := bufferedReader.ReadString('\n')
		if errorReading3 != nil && errorReading3 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading3)
			return
		}
		bothConditionsPopulation = strings.TrimSpace(bothConditionsPopulation)
		slicePopulation := strings.Split(bothConditionsPopulation, ",")

		if len(slicePopulation) != 2 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 2 parameters.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawPopulationSize, errorTransforming9 := strconv.ParseInt(strings.TrimSpace(slicePopulation[0]), 10, 64)
		if errorTransforming9 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Population size: %v", errorTransforming9)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		rawGenerations, errorTransforming10 := strconv.ParseInt(strings.TrimSpace(slicePopulation[1]), 10, 64)
		if errorTransforming10 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Total generations: %v", errorTransforming10)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if rawPopulationSize > 1 && rawGenerations > 0 {
			fmt.Println("✨ [SUCCESS] Evolution parameters successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			populationSize = int(rawPopulationSize)
			generations = int(rawGenerations)
			break
		} else {
			fmt.Printf("⚠️ [VALIDATION FAILED] Invalid constraints! Size (%v) must be > 1, Generations (%v) must be > 0.\n", populationSize, generations)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("🚀 INITIALIZING GENETIC SIMULATION ENGINE... 🚀")
	fmt.Println("==================================================")
	fmt.Printf("🎯 Target Lock:     X=%d, Y=%d, Z=%d\n", targetCoordinateX, targetCoordinateY, targetCoordinateZ)
	fmt.Printf("👥 Population:    %d specimens | %d generations\n", populationSize, generations)
	fmt.Printf("🧬 First DNA:       Pitch=%.2f°, Fuel=%.2f kg, Burn=%.2f kg/s\n", pitchFirstRocket, fuelFirstRocket, burnRateFirstRocket)
	fmt.Printf("==================================================\n")

	calculatedEngineEfficiency := 800.0 + (burnRateFirstRocket/bodyMassPopulation)*1500.0

	firstRocket := simulation.Rocket{PitchDegree: pitchFirstRocket, YawDegree: yawFirstRocket, Fuel: fuelFirstRocket, BurnRate: burnRateFirstRocket, EngineEfficiency: calculatedEngineEfficiency, BodyMass: bodyMassPopulation}
	LaunchPopulation(firstRocket, int(targetCoordinateX), int(targetCoordinateY), int(targetCoordinateZ), int(populationSize), int(generations))
}
