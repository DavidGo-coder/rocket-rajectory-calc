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

func LaunchPopulation(FirstRocket simulation.Rocket, TargetCoordinateX, TargetCoordinateY, TargetCoordinateZ, PopulationSize, Generations int, BodyMassPopulation float64) {
	Listener, errorListening := net.Listen("tcp", ":8080")
	if errorListening != nil {
		log.Fatalf("❌ Failed to bind TCP port 8080: %v\n", errorListening)
	}
	defer Listener.Close()

	fmt.Println("📡 [NET SYSTEM] TCP Server started successfully on port 8080")

	Connection, errorConection := Listener.Accept()
	if errorConection != nil {
		log.Printf("❌ Python client handshake failed: %v", errorConection)
		return
	}
	defer Connection.Close()

	fmt.Println("✨ [CONNECTION] Python neural-radar pipeline connected!")

	BothTargetPacket := fmt.Sprintf("TARGET|%d,%d\n", TargetCoordinateX, TargetCoordinateY)
	Connection.Write([]byte(BothTargetPacket))
	time.Sleep(150 * time.Millisecond)

	RocketPopulation := genetic.FirstGeneticCalculation(FirstRocket, PopulationSize, BodyMassPopulation)

	SliceTenBestRocket, MedianPopulation := simulation.CalculationTrajectory(TargetCoordinateX, TargetCoordinateY, TargetCoordinateZ, PopulationSize, BodyMassPopulation, RocketPopulation)
	InitialMedian := MedianPopulation

	for id, rocket := range SliceTenBestRocket {
		DataCSV := fmt.Sprintf("ROCKET_ID:%d;%s\n", id+1, rocket.TrajectoryCSV)
		Connection.Write([]byte(DataCSV))
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(3000 * time.Millisecond)

	RocketPopulation2 := genetic.GeneticCalculation(SliceTenBestRocket, PopulationSize, MedianPopulation, BodyMassPopulation)

	for CountGenerations := 0; CountGenerations < Generations; CountGenerations++ {
		fmt.Printf("⏳ [EVOLUTION] Simulation of generation %d/%d running...\n", CountGenerations+1, Generations)
		SliceTenBestRocket2, MedianPopulation2 := simulation.CalculationTrajectory(TargetCoordinateX, TargetCoordinateY, TargetCoordinateZ, PopulationSize, BodyMassPopulation, RocketPopulation2)

		AdaptiveThreshold := math.Max(80.0, InitialMedian*0.025)

		for id, Rocket := range SliceTenBestRocket2 {
			if Rocket.TrajectoryCSV == "" {
				Rocket.TrajectoryCSV = "0.00,0.00"
			}

			DataCSV := fmt.Sprintf("ROCKET_ID:%d;%s\n", id+1, Rocket.TrajectoryCSV)

			_, errorWriting2 := Connection.Write([]byte(DataCSV))
			if errorWriting2 != nil && errorWriting2 != io.EOF {
				return
			}

			if Rocket.Fitness <= AdaptiveThreshold && Rocket.X > 10 {
				fmt.Println("\n==================================================")
				fmt.Println("🎯 THE GOAL WAS ACHIEVED! WE HAVE A WINNER! 🎯")
				fmt.Println("==================================================")
				fmt.Printf("🚀 Best Rocket Characteristics:\n")
				fmt.Printf("   • Pitch Angle:   %.2f degrees\n", Rocket.PitchDegree)
				fmt.Printf("   • Initial Fuel:  %.2f kg\n", Rocket.Fuel)
				fmt.Printf("   • Engine Burn:   %.2f kg/s\n", Rocket.BurnRate)
				fmt.Printf("--------------------------------------------------\n")
				fmt.Printf("📊 Flight Performance:\n")
				fmt.Printf("   • Final X Pos:   %.2f meters\n", Rocket.X)
				fmt.Printf("   • Max Y Pos:  %.2f meters (Max Y)\n", Rocket.Y)
				fmt.Printf("   • Target Miss:   %.2f meters\n", Rocket.Fitness)
				fmt.Println("==================================================")

				fmt.Println("\n👋 [SYSTEM] Simulation successfully completed. Closing connections...")
				time.Sleep(1 * time.Second)
				return
			}
			time.Sleep(20 * time.Millisecond)
		}

		time.Sleep(3000 * time.Millisecond)
		RocketPopulation2 = genetic.GeneticCalculation(SliceTenBestRocket2, PopulationSize, MedianPopulation2, BodyMassPopulation)
	}
}

func main() {
	BufferedReader := bufio.NewReader(os.Stdin)

	TargetCoordinateX := 0
	TargetCoordinateY := 0
	TargetCoordinateZ := 0
	PitchFirstRocket := 0.0
	YawFirstRocket := 0.0
	FuelFirstRocket := 0.0
	BurnRateFirstRocket := 0.0
	BodyMassPopulation := 0.0
	PopulationSize := 0
	Generations := 0

	for {
		fmt.Println("┌──────────────────────────────────────────────────┐")
		fmt.Println("│ 🎯 TARGET COORDINATES SPECIFICATION              │")
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println(" Enter position in meters using format: X, Y, Z")
		fmt.Print(" ❯ ")

		BothTargetCoordinates, errorReading1 := BufferedReader.ReadString('\n')
		if errorReading1 != nil && errorReading1 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading1)
			return
		}
		BothTargetCoordinates = strings.TrimSpace(BothTargetCoordinates)
		TargetSlice := strings.Split(BothTargetCoordinates, ",")

		if len(TargetSlice) != 3 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 3 coordinates (X,Y,Z).")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawTargetCoordinateX, errorTransforming1 := strconv.ParseInt(strings.TrimSpace(TargetSlice[0]), 10, 64)
		if errorTransforming1 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse X: %v", errorTransforming1)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawTargetCoordinateY, errorTransforming2 := strconv.ParseInt(strings.TrimSpace(TargetSlice[1]), 10, 64)
		if errorTransforming2 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Y: %v", errorTransforming2)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawTargetCoordinateZ, errorTransforming3 := strconv.ParseInt(strings.TrimSpace(TargetSlice[2]), 10, 64)
		if errorTransforming3 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Z: %v", errorTransforming3)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if RawTargetCoordinateX > 0 && RawTargetCoordinateY >= 0 && RawTargetCoordinateZ > 0 {
			fmt.Println("✨ [SUCCESS] Target coordinates successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			TargetCoordinateX = int(RawTargetCoordinateX)
			TargetCoordinateY = int(RawTargetCoordinateY)
			TargetCoordinateZ = int(RawTargetCoordinateZ)
			break
		} else {
			fmt.Println("⚠️ [VALIDATION FAILED] Coordinates out of bounds! X, Y, and Z must be positive values.")
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

		BothConditionsFirstRocket, errorReading2 := BufferedReader.ReadString('\n')
		if errorReading2 != nil && errorReading2 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading2)
			return
		}
		BothConditionsFirstRocket = strings.TrimSpace(BothConditionsFirstRocket)
		SliceConditionsFirstRocket := strings.Split(BothConditionsFirstRocket, ",")

		if len(SliceConditionsFirstRocket) != 5 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 5 parameters.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawPitchFirstRocket, errorTransforming4 := strconv.ParseFloat(strings.TrimSpace(SliceConditionsFirstRocket[0]), 64)
		if errorTransforming4 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Pitch: %v", errorTransforming4)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawYawFirstRocket, errorTransforming5 := strconv.ParseFloat(strings.TrimSpace(SliceConditionsFirstRocket[1]), 64)
		if errorTransforming5 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Yaw: %v", errorTransforming5)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawFuelFirstRocket, errorTransforming6 := strconv.ParseFloat(strings.TrimSpace(SliceConditionsFirstRocket[2]), 64)
		if errorTransforming6 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Fuel mass: %v", errorTransforming6)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawBurnRateFirstRocket, errorTransforming7 := strconv.ParseFloat(strings.TrimSpace(SliceConditionsFirstRocket[3]), 64)
		if errorTransforming7 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Burn rate: %v", errorTransforming7)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawBodyMassPopulation, errorTransforming8 := strconv.ParseFloat(strings.TrimSpace(SliceConditionsFirstRocket[4]), 64)
		if errorTransforming8 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Body mass: %v", errorTransforming8)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if RawYawFirstRocket > 1 && RawYawFirstRocket < 89 && RawPitchFirstRocket > 1 && RawPitchFirstRocket < 89 && RawFuelFirstRocket > 0 && RawBurnRateFirstRocket > 0 && RawBodyMassPopulation > 0 {
			fmt.Println("✨ [SUCCESS] Rocket parameters successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			PitchFirstRocket = RawPitchFirstRocket
			YawFirstRocket = RawYawFirstRocket
			FuelFirstRocket = RawFuelFirstRocket
			BurnRateFirstRocket = RawBurnRateFirstRocket
			BodyMassPopulation = RawBodyMassPopulation
			break
		} else {
			fmt.Println("⚠️ [VALIDATION FAILED] Out of range! Angles must be 1-89°, other metrics must be positive.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}
	}

	for {
		fmt.Println("\n┌──────────────────────────────────────────────────┐")
		fmt.Println("│ 👥 EVOLUTIONARY POPULATION CONSTRAINTS           │")
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println(" Enter parameters separated by commas:")
		fmt.Println(" PopulationSize (min: 11), TotalGenerations")
		fmt.Print(" ❯ ")

		BothConditionsPopulation, errorReading3 := BufferedReader.ReadString('\n')
		if errorReading3 != nil && errorReading3 != io.EOF {
			log.Printf("⚠️ Critical IO Error: %v", errorReading3)
			return
		}
		BothConditionsPopulation = strings.TrimSpace(BothConditionsPopulation)
		SlicePopulation := strings.Split(BothConditionsPopulation, ",")

		if len(SlicePopulation) != 2 {
			fmt.Println("❌ [INPUT ERROR] Invalid format! Please enter exactly 2 parameters.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawPopulationSize, errorTransforming9 := strconv.ParseInt(strings.TrimSpace(SlicePopulation[0]), 10, 64)
		if errorTransforming9 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Population size: %v", errorTransforming9)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		RawGenerations, errorTransforming10 := strconv.ParseInt(strings.TrimSpace(SlicePopulation[1]), 10, 64)
		if errorTransforming10 != nil {
			log.Printf("❌ [PARSING ERROR] Failed to parse Total generations: %v", errorTransforming10)
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}

		if RawPopulationSize > 10 && RawGenerations > 0 {
			fmt.Println("✨ [SUCCESS] Evolution parameters successfully validated.")
			fmt.Println("──────────────────────────────────────────────────")
			PopulationSize = int(RawPopulationSize)
			Generations = int(RawGenerations)
			break
		} else {
			fmt.Println("⚠️ [VALIDATION FAILED] Invalid constraints! Size must be > 10, Generations must be > 0.")
			fmt.Println("──────────────────────────────────────────────────")
			continue
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("🚀 INITIALIZING GENETIC SIMULATION ENGINE... 🚀")
	fmt.Println("==================================================")
	fmt.Printf("🎯 Target Lock:     X=%d, Y=%d, Z=%d\n", TargetCoordinateX, TargetCoordinateY, TargetCoordinateZ)
	fmt.Printf("👥 Population:    %d specimens | %d generations\n", PopulationSize, Generations)
	fmt.Printf("🧬 First DNA:       Pitch=%.2f°, Fuel=%.2f kg, Burn=%.2f kg/s\n", PitchFirstRocket, FuelFirstRocket, BurnRateFirstRocket)
	fmt.Printf("==================================================\n")

	CalculatedEngineEfficiency := 800.0 + (BurnRateFirstRocket/BodyMassPopulation)*1500.0

	FirstRocket := simulation.Rocket{PitchDegree: PitchFirstRocket, YawDegree: YawFirstRocket, Fuel: FuelFirstRocket, BurnRate: BurnRateFirstRocket, EngineEfficiency: CalculatedEngineEfficiency}
	LaunchPopulation(FirstRocket, int(TargetCoordinateX), int(TargetCoordinateY), int(TargetCoordinateZ), int(PopulationSize), int(Generations), BodyMassPopulation)
}
