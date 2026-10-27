# 🚀 Rocket Trajectory Distributed Genetic Optimization Cluster (v2.0.0)

A high-performance, concurrent **distributed evolutionary simulation complex** designed to optimize rocket launch parameters. The ecosystem has been completely re-engineered from a monolithic design into a **mesh of 5 independent nodes (microservices)** communicating over serialized real-time TCP sockets.

---

## 🏗️ Distributed Cluster Topology

The system enforces strict separation of concerns via a decentralized layout. Every component operates as an isolated network node, linking to a central Master Orchestrator:

```text
       [ input ] (CLI Input Utility)
           │ (Initial Session Profiles)
           ▼
     [   master   ] <───(TCP:8800)───> [ genetic ] (Evolutionary GA Engine)
  (Network Orchestrator)
           │ 
           ├───(TCP:8800)───> [ simulation ] (Concurrent Physics Processor)
           │
           └───(TCP:8800)───> [ visualize.py ] (Asynchronous Pygame Radar Canvas)
```

### 🛰️ The 5 Core Network Nodes:
1. **`master` (Go):** The central hub and network router. Manages data transmission phases, handles frame-rate pacing (`waitSlowPython`), orchestrates socket lifecycles via LIFO `defer` execution, and gracefully terminates the session using the `FINISHWORKING` signal.
2. **`input` (Go):** The client-facing mission profiling interface. Validates 3D target points (`X, Y, Z`), structural rocket seeds, and population constraints. It pipes this metadata payload directly into the master and disconnects immediately.
3. **`simulation` (Go):** The parallelized heavy-computing physics node. Spawns isolated Go-routines for every individual chromosome via `sync.WaitGroup`. Integrates trajectories using Euler-Cromer equations and streams an 11-parameter telemetry chunk back to the master.
4. **`genetic` (Go):** The evolutionary calculation server. Maintains a rolling generation score buffer to detect optimization stagnation, triggers gene crossovers/mutations, and safely boots adaptive optimization routines starting at Gen 4 (`if countGenerations >= 4`).
5. **`visualize.py` (Python):** The asynchronous telemetry radar driven by Pygame and Rich. Configured with non-blocking reads (`client.setblocking(False)`), protected against TCP packet fragmentation via a string line-buffer, and renders the top 10 trajectories of each generation instantly.

---

## 🏎️ Core Telemetry & Network Protocol

### 1. Flight Dynamics Simulation (Euler-Cromer Integrator)
Flight profiles are processed utilizing a discrete time-step of Δ t = 0.01 seconds. The physics matrix tracks:
* **Exponential Atmospheric Density Decay:** Aerodynamic drag forces drop dynamically with altitude:
  \[F_{\text{drag}} = \text{AirResistance} \cdot e^{-\frac{y}{8500}} \cdot v^2\]
* **Newton's Inverse-Square Gravity Decay:** Earth's gravitational acceleration drops as the rocket climbs:
  \[g(y) = g_0 \cdot \left(\frac{R_E}{R_E + y}\right)^2\]
* **Failed Launch Penalty:** Specimens failing to clear an apogee of **5.0 meters** are automatically penalized with a flat **+50,000 meter** penalty added to their final miss distance (`FailedLaunchPenalty`).

### 2. TCP Data Layout (11-Parameter Serialization Packet)
Data streaming between the Simulation component and the Master server utilizes a highly strict index-based mapping using the `|` delimiter:
```text
Index:  [0]   [1]   [2]    [3]        [4]       [5]        [6]  [7]  [8]  [9]     [10]
Fields: Pitch|Yaw  |Fuel  |BurnRate  |BodyMass |Efficiency|XPos|YPos|ZPos|Fitness|TrajectoryCSV
```
*The master extracts the trailing `TrajectoryCSV` string and forwards it to the Pygame radar using the following data signature: `ROCKET_ID:id;x1,y1;x2,y2;...`.*

---

## ⚙️ Core Cluster Configuration Constants

| Constant | Value | Description |
| :--- | :--- | :--- |
| `portConnection` | `8800` | Dedicated cluster TCP socket port handle |
| `timeOneGeneration` | `3s` | Master synchronization pacing window for smooth GUI display |
| `medianHistorySize` | `5` | Sliding history matrix size to analyze population stagnation |
| `OneMomentSimulation`| `0.01s` | Euler-Cromer integration physics delta step ($\Delta t$) |
| `FailedLaunchPenalty`| `50000.0` | Target miss penalty added for failed lift-offs |

---

## 🏁 Cluster Bootstrapping Sequence

To guarantee that all socket tunnels establish handshakes smoothly without throwing abrupt `io.EOF` disconnect errors, you **must** launch all 5 terminals in the exact order below.

### Step 1: Fire up the Orchestrator
Open your first terminal window and activate the central server routing node:
```bash
cd master
go run .
```

### Step 2: Initialize the Telemetry Radar Canvas
In a secondary shell, boot the Pygame graphics client to pass the `I|AM|VISUALIZATION` network validation phase:
```bash
python visualize.py
```

### Step 3: Attach the Computational Physics Engine
In a third terminal window, launch the high-concurrency simulation service:
```bash
cd simulation
go run .
```

### Step 4: Boot the Genetic Processing Node
In a fourth shell, bring up the evolutionary operations engine:
```bash
cd genetic
go run .
```

### Step 5: Execute the Mission Launch (The Trigger)
In your final terminal window, run the CLI parsing tool. Input your target coordinates and metrics to fire up the simulation loop:
```bash
cd input
go run .
```

Once parameters are defined within the `input` shell, the cluster will automatically crunch numbers for $N$ generations. Upon hitting the **Adaptive Threshold Limit** ($\max(80.0, \text{InitialMedian} \times 0.005)$), the Master gracefully distributes the `FINISHWORKING` signal, closing all terminal processes safely.
