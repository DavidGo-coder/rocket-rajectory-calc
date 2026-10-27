# 🪐 Rocket Distributed Genetic Optimization Cluster (v3.0.0)

A high-performance, concurrent, distributed evolutionary simulation complex designed to optimize rocket launch parameters. The ecosystem is built on top of a multi-threaded Go orchestration engine and an asynchronous Python (Pygame + Rich) graphical client, communicating seamlessly over serialized real-time TCP sockets via local port `8800`.

---

## 🏗️ Distributed Cluster Topology

The system enforces strict separation of concerns via decentralized logical layouts. Instead of demanding five separate standalone programs, the architecture is elegantly compressed into **3 independent running processes**, where the single Python client dynamically masquerades as **3 distinctive network profiles** by spawning separate background execution threads.

```text
     [  visualize.py  ] (Python Client: GUI, Real-Time Logging, Config Injection)
        ├── Thread 1 (PARSER) ───> [   master   ] <───> [ genetic ] (Evolutionary GA Server)
        ├── Thread 2 (PRINTER) ──> (Orchestrator)
        └── Thread 3 (VISUAL) ───>       │
                                         └───> [ simulation ] (Concurrent Physics Engine)
```

### 🛰️ Core Network Nodes & Active Processes:

1. **`master` (Go):** The central hub and cluster network router. Manages frame-rate pacing sequences (`waitSlowPython`), orchestrates multi-node data transmission phases, handles socket lifecycles, and broadcasts orderly shutdowns via the `FINISHWORKING` packet signal.
2. **`simulation` (Go):** The heavy-computing parallel physics engine. Spawns isolated Go-routines for every individual specimen via `sync.WaitGroup`, integrating trajectories dynamically based on the properties of the selected celestial body.
3. **`genetic` (Go):** The evolutionary calculation server. Maintains generation fitness history scoring, triggers gene crossovers/mutations, and forces emergency population diversification (shuffling a fresh random vector field) to shatter optimization stagnation.
4. **`visualize.py` (Python):** A multi-role graphical powerhouse client that establishes connection handshakes with the orchestrator under **3 different structural identities**:
   * **`I|AM|PARSER`** — The config injector thread. Extracts parameters out of the live GUI fields and flashes the raw data matrix across the tunnel once the launch button is pressed.
   * **`I|AM|PRINTER`** — The logging thread. Intercepts generation medians to build live, responsive progress bars right in your shell window utilizing the `Rich` console library, finishing with the final optimized champion card.
   * **`I|AM|VISUALIZATION`** — The 3D radar canvas thread. Pulls compressed arrays of flight trajectories and renders them on the Pygame interface with interactive mouse dragging for Pitch/Yaw and canvas zoom controls.

---

## 🚀 Solar System Physics Framework

The physics integration engine is decoupled from hardcoded Earth values. All celestial constants are isolated into a highly scalable `PlanetProps` structure array, fetched instantly with an O(1) lookup via a hash map:

```go
type PlanetProps struct {
    Name             string
    Gravity          float64 // Base surface gravitational acceleration (g0)
    Radius           float64 // Volumetric mean planet radius (Re)
    AirDensityCoeff  float64 // Atmospheric scaling multiplier coefficient
}
```

### 🌌 Available Celestial Object Registry:
*   **EARTH:** `Gravity: 9.81`, `Radius: 6371000.0`, `AirDensityCoeff: 1.0` (The baseline reference model)
*   **MARS:** `Gravity: 3.71`, `Radius: 3389500.0`, `AirDensityCoeff: 0.01` (Thin, hyper-low-density atmosphere)
*   **MOON:** `Gravity: 1.62`, `Radius: 1737400.0`, `AirDensityCoeff: 0.0` (Complete absolute vacuum environment)
*   **VENUS:** `Gravity: 8.87`, `Radius: 6051800.0`, `AirDensityCoeff: 90.0` (Super-dense, high-pressure atmosphere)

### 📈 Flight Dynamics Integration (Euler-Cromer Equations):
*   **Inverse-Square Gravity Decay:** Gravitational pulling forces drop off dynamically along with altitude changes (y) relative to the planetary core:
    \[g(y) = g_0 \cdot \left(\frac{R_E}{R_E + y}\right)^2\]
*   **Exponential Drag Decay:** Aerodynamic drag drops off exponentially in higher atmospheric layers and scales directly with the chosen planetary density constant:
    \[F_{\text{drag}} = \text{AirResistance} \cdot \text{planet.AirDensityCoeff} \cdot e^{-\frac{y}{8500}} \cdot v^2\]
*   **Failed Launch Penalty (`FailedLaunchPenalty`):** Specimens that fail to clear a threshold altitude of **5.0 meters** are slapped with a heavy **+50,000 meter** penalty appended directly to their final miss metric (`Fitness`).

---

## 🏎️ Telemetry Data Layout (11-Parameter Packet)

Streaming pipelines between the physics module (`simulation`) and the cluster router (`master`) pass flat delimited string records using the `|` marker:

```text
Index:  [0]   [1]   [2]    [3]        [4]       [5]        [6]    [7]    [8]    [9]     [10]
Fields: Pitch|Yaw  |Fuel  |BurnRate  |BodyMass |Efficiency|XPos  |YPos  |ZPos  |Fitness|TrajectoryCSV
```

The trailing `TrajectoryCSV` contains the packed vector path of the rocket flight profile coordinates (`x1,y1,z1;x2,y2,z2;...`), which the `master` routes directly to the Pygame radar thread for immediate drawing.

---

## 🏁 Cluster Bootstrapping Sequence

To guarantee that all socket tunnels establish handshakes smoothly without throwing abrupt `io.EOF` disconnect errors, you **must** launch your terminal processes inside VS Code in this exact sequence:

### Step 1: Spin up the Network Orchestrator
Open your first terminal layout and fire up the main координирующий hub server to begin listening for cluster nodes:
```bash
cd master
go run .
```

### Step 2: Attach the Heavy-Computing Physics Unit
In your second terminal tab, spin up the parallelized calculation matrix engine:
```bash
cd simulation
go run .
```

### Step 3: Run the Evolutionary Genetic Core
In a third shell terminal, bring up the DNA crossover and mutation server:
```bash
cd genetic
go run .
```

### Step 4: Boot the Multi-Threaded Python Interface
In your final terminal, trigger the script. It will spawn its background networking worker threads and attach to the orchestrator simultaneously as `PARSER`, `PRINTER`, and `VISUALIZATION`:
```bash
python visualize.py
```

**Igniting the Simulation:** Modify values inside the live input fields on the bottom panel of the Pygame UI, and hit the **🔥 INJECT CONFIG & IGNITE ENGINE** button. The cluster will spin up natural selection algorithms in a loop. Once it hits the targeted precision ceiling, the orchestrator triggers an automatic cluster-wide cleanup signal.
