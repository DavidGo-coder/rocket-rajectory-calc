package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	core "github.com/DavidGo-coder/rocket-trajecory-calc"
)

const (
	portConnection string = "8800"
	hostConnection string = "127.0.0.1"
	colorReset            = "\x1b[0m"
	colorRed              = "\x1b[31m"
	colorGreen            = "\x1b[32m"
	colorYellow           = "\x1b[33m"
	colorBlue             = "\x1b[34m"
	colorPurple           = "\x1b[35m"
	colorCyan             = "\x1b[36m"
	colorBold             = "\x1b[1m"
)

func readTarget(bufferedReader *bufio.Reader) (targetX, targetY, targetZ int, err error) {
	for {
		fmt.Printf("\n🎯 %s[TARGET SPECIFICATION]%s Enter position in meters (Format: X, Y, Z)\n", colorCyan, colorReset)
		fmt.Printf(" %s❯%s ", colorCyan, colorReset)

		targets, err := core.ReadInputCommaSliced(bufferedReader)
		if err != nil {
			log.Printf("⚠️  Critical IO Error: %v", err)
			return targetX, targetY, targetZ, err
		}

		if len(targets) != 3 {
			fmt.Printf("❌ %s[INPUT ERROR]%s Invalid format! Please enter exactly 3 coordinates (X,Y,Z).\n", colorRed, colorReset)
			continue
		}

		rawX, err := core.ParseInt(targets, 0)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse X: %v\n", colorRed, colorReset, err)
			continue
		}

		rawY, err := core.ParseInt(targets, 1)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Y: %v\n", colorRed, colorReset, err)
			continue
		}

		rawZ, err := core.ParseInt(targets, 2)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Z: %v\n", colorRed, colorReset, err)
			continue
		}

		if (rawX > 0 && rawY >= 0 && rawZ >= 0) || (rawX >= 0 && rawY >= 0 && rawZ > 0) {
			fmt.Printf("✨ %s[SUCCESS]%s Target coordinates successfully validated.\n", colorGreen, colorReset)
			targetX = rawX
			targetY = rawY
			targetZ = rawZ
			break
		} else {
			fmt.Printf("⚠️  %s[VALIDATION FAILED]%s At least X or Z must be positive. Y cannot be lonely.\n", colorYellow, colorReset)
			fmt.Printf("    Got: X=%v, Y=%v, Z=%v\n", rawX, rawY, rawZ)
			continue
		}
	}
	return targetX, targetY, targetZ, err
}

func readFirstRocket(bufferedReader *bufio.Reader) (rocket core.Rocket, err error) {
	pitch, yaw, fuel, burnRate, bodyMass := 0.0, 0.0, 0.0, 0.0, 0.0

	for {
		fmt.Printf("\n🚀 %s[CORE DNA]%s Enter 5 metrics (Pitch, Yaw, Fuel, BurnRate, BodyMass)\n", colorPurple, colorReset)
		fmt.Printf(" %s❯%s ", colorPurple, colorReset)

		parameters, err := core.ReadInputCommaSliced(bufferedReader)
		if err != nil {
			log.Printf("⚠️  Critical IO Error: %v", err)
			return rocket, err
		}

		if len(parameters) != 5 {
			fmt.Printf("❌ %s[INPUT ERROR]%s Please enter exactly 5 parameters.\n", colorRed, colorReset)
			continue
		}

		rawPitchRocket, err := core.ParseFloat(parameters, 0)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Pitch: %v\n", colorRed, colorReset, err)
			continue
		}

		rawYawRocket, err := core.ParseFloat(parameters, 1)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Yaw: %v\n", colorRed, colorReset, err)
			continue
		}

		rawFuelRocket, err := core.ParseFloat(parameters, 2)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Fuel: %v\n", colorRed, colorReset, err)
			continue
		}

		rawBurnRateRocket, err := core.ParseFloat(parameters, 3)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse BurnRate: %v\n", colorRed, colorReset, err)
			continue
		}

		rawBodyMassRocket, err := core.ParseFloat(parameters, 4)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse BodyMass: %v\n", colorRed, colorReset, err)
			continue
		}

		if rawYawRocket >= 1 && rawYawRocket <= 89 && rawPitchRocket >= 1 && rawPitchRocket <= 89 && rawFuelRocket > 0 && rawBurnRateRocket > 0 && rawBodyMassRocket > 0 {
			fmt.Printf("✨ %s[SUCCESS]%s Rocket DNA matches configuration constraints.\n", colorGreen, colorReset)
			pitch = rawPitchRocket
			yaw = rawYawRocket
			fuel = rawFuelRocket
			burnRate = rawBurnRateRocket
			bodyMass = rawBodyMassRocket
			break
		} else {
			fmt.Printf("⚠️  %s[VALIDATION FAILED]%s Angles (%v°, %v°) must be 1-89°, metrics must be positive.\n", colorYellow, colorReset, rawPitchRocket, rawYawRocket)
			continue
		}
	}

	calculatedEngineEfficiency := 800.0 + (burnRate/bodyMass)*1500.0
	rocket = core.Rocket{PitchDegree: pitch, YawDegree: yaw, Fuel: fuel, BurnRate: burnRate, EngineEfficiency: calculatedEngineEfficiency, BodyMass: bodyMass}

	return rocket, err
}

func readPopulationParameters(bufferedReader *bufio.Reader) (generations, populationSize int, err error) {
	for {
		fmt.Printf("\n👥 %s[EVOLUTION CONSTRAINTS]%s Enter: PopulationSize, TotalGenerations\n", colorBlue, colorReset)
		fmt.Printf(" %s❯%s ", colorBlue, colorReset)

		conditions, err := core.ReadInputCommaSliced(bufferedReader)
		if err != nil {
			log.Printf("⚠️  Critical IO Error: %v", err)
			return generations, populationSize, err
		}

		if len(conditions) != 2 {
			fmt.Printf("❌ %s[INPUT ERROR]%s Expected exactly 2 parameters.\n", colorRed, colorReset)
			continue
		}

		rawPopSize, err := core.ParseInt(conditions, 0)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse PopSize: %v\n", colorRed, colorReset, err)
			continue
		}

		rawGenerations, err := core.ParseInt(conditions, 1)
		if err != nil {
			fmt.Printf("❌ %s[PARSING ERROR]%s Failed to parse Generations: %v\n", colorRed, colorReset, err)
			continue
		}

		if rawPopSize > 1 && rawGenerations > 0 {
			fmt.Printf("✨ %s[SUCCESS]%s Evolution criteria confirmed.\n", colorGreen, colorReset)
			populationSize = rawPopSize
			generations = rawGenerations
			break
		} else {
			fmt.Printf("⚠️  %s[VALIDATION FAILED]%s Size (%v) must be > 1, Generations (%v) must be > 0.\n", colorYellow, colorReset, rawPopSize, rawGenerations)
			continue
		}
	}
	return generations, populationSize, err
}

func readAllInput() (targetX, targetY, targetZ int, rocket core.Rocket, populationSize, generations int, err error) {
	bufferedReader := bufio.NewReader(os.Stdin)

	targetX, targetY, targetZ, err = readTarget(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generations, err
	}

	firstRocket, err := readFirstRocket(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generations, err
	}

	generations, populationSize, err = readPopulationParameters(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, rocket, populationSize, generations, err
	}

	fmt.Println("\n" + colorCyan + "⚡ ══════════════════ SIMULATION PROFILE ══════════════════ ⚡" + colorReset)
	fmt.Printf(" 🎯 Target coordinates :  X=%v, Y=%v, Z=%v\n", targetX, targetY, targetZ)
	fmt.Printf(" 👥 Engine parameters   :  %v specimens | %v generations\n", populationSize, generations)
	fmt.Printf(" 🧬 Core DNA profile   :  Pitch=%.2f° | Yaw=%.2f° | Fuel=%.1f kg\n", firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel)
	fmt.Printf(" 🛠️  Physical baseline  :  Mass=%.1f kg | Burn=%.1f kg/s | Efficiency=%.1f\n", firstRocket.BodyMass, firstRocket.BurnRate, firstRocket.EngineEfficiency)
	fmt.Println(colorCyan + "══════════════════════════════════════════════════════════" + colorReset)

	return targetX, targetY, targetZ, firstRocket, populationSize, generations, nil
}

func main() {
	targetX, targetY, targetZ, rocket, populationSize, generations, errReading := readAllInput()
	if errReading != nil {
		fmt.Printf("\n❌ %s%s[FATAL INITIALIZATION ERROR]%s %v\n", colorRed, colorBold, colorReset, errReading)
		return
	}

	address := net.JoinHostPort(hostConnection, portConnection)

	fmt.Printf("🌐 %s[NETWORK]%s Connecting to simulation host at %s...\n", colorCyan, colorReset, address)
	connection, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ %s%s[CONNECTION FAILED]%s Cannot establish stream tunnel: %v\n", colorRed, colorBold, colorReset, err)
		return
	}
	defer connection.Close()
	fmt.Printf("✅ %s[CONNECTED]%s Connected to physics cluster.\n", colorGreen, colorReset)

	writer := bufio.NewWriter(connection)
	defer func() {
		if err := writer.Flush(); err != nil {
			fmt.Printf("❌ %s[IO FLUSH ERROR]%s Data loss alert! Buffer dropped: %v\n", colorRed, colorReset, err)
		} else {
			fmt.Printf("✨ %s[SUCCESS]%s Data pipeline synchronized. Packets delivered successfully.\n", colorGreen, colorReset)
		}
	}()

	fmt.Printf("📤 %s[TRANSMISSION]%s Streaming configuration matrices to remote interpreter...\n", colorBlue, colorReset)

	if _, err = fmt.Fprintln(writer, "I|AM|PARSER"); err != nil {
		fmt.Printf("💥 %s[HANDSHAKE FAILED]%s Packet rejected by backend server: %v\n", colorRed, colorReset, err)
		return
	}

	if _, err = fmt.Fprintf(writer, "%v|%v|%v\n", targetX, targetY, targetZ); err != nil {
		fmt.Printf("💥 %s[TRANSMISSION CRASH]%s Target coordinate array drop: %v\n", colorRed, colorReset, err)
		return
	}

	if _, err = fmt.Fprintf(writer, "%v|%v|%v|%v|%v|%v\n", rocket.PitchDegree, rocket.YawDegree, rocket.Fuel, rocket.BurnRate, rocket.BodyMass, rocket.EngineEfficiency); err != nil {
		fmt.Printf("💥 %s[TRANSMISSION CRASH]%s Core rocket structural payload drop: %v\n", colorRed, colorReset, err)
		return
	}

	if _, err = fmt.Fprintf(writer, "%v|%v\n", populationSize, generations); err != nil {
		fmt.Printf("💥 %s[TRANSMISSION CRASH]%s Generation boundaries matrix drop: %v\n", colorRed, colorReset, err)
		return
	}

	writer.Flush() // Форсируем отправку всех пакетов сразу
	fmt.Printf("📦 %s[PACKET SENT]%s All DNA structures successfully queued for calculation.\n", colorGreen, colorReset)
}
