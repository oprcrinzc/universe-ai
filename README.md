<div align="center">

# 🪐 GRAVITY MASS SIMULATOR 3D
### *Interactive 3D N-Body Orbital Physics Sandbox in Go & Raylib*

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Engine](https://img.shields.io/badge/Engine-Raylib--Go_v6.0-black?style=for-the-badge&logo=c&logoColor=red)](https://github.com/gen2brain/raylib-go)
[![Platform](https://img.shields.io/badge/Platform-Linux_%7C_Windows_%7C_macOS-blue?style=for-the-badge&logo=linux&logoColor=white)](https://github.com)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Physics](https://img.shields.io/badge/Physics-Velocity_Verlet_Symplectic-orange?style=for-the-badge)](https://en.wikipedia.org/wiki/Verlet_integration)
[![Performance](https://img.shields.io/badge/Performance-60_FPS_%7C_Sub--Stepped-brightgreen?style=for-the-badge)](https://github.com)

<p align="center">
  <b>Simulate chaotic solar systems, binary star choreography, supermassive accretion disks, and deep-space gravitational dances in real time with hardware-accelerated 3D graphics and flight telemetry.</b>
</p>

[Video Showcase](#-video-showcase) • [Key Features](#-key-features) • [Preset Scenarios](#-preset-scenarios) • [Screenshot Gallery](#-screenshot-gallery) • [Gameplay & Flight Computer](#-gameplay--flight-computer) • [Controls](#-controls--hotkeys) • [Platform Support](#-supported-platforms) • [Quick Start](#-installation--quick-start) • [Physics Model](#-mathematical--physics-engine)

---

### 🎬 Real-Time Motion Preview
![Gravity Simulator 3D Showcase Preview](assets/showcase_preview.gif)

*60-body Keplerian accretion disk swirling around a central supermassive gravitational singularity.*

</div>

---

## 📽️ Video Showcase

Experience silky-smooth orbital mechanics and dynamic trail fading in action:

| High-Definition Video Clip | Formats Available |
| :--- | :--- |
| 📹 **[Watch Video Showcase (MP4)](assets/showcase_web.mp4)** | `MP4 (H.264 / 1080p Web-Optimized)` • `assets/showcase_web.mp4` |
| 🎞️ **High-Res Master Video** | `assets/showcase.mp4` |
| 🎞️ **Animated Showcase GIF** | `assets/showcase_preview.gif` (3.2 MB) |

> 💡 **Tip:** You can download or stream [`assets/showcase_web.mp4`](assets/showcase_web.mp4) locally or directly on GitHub to view full 60 FPS celestial interactions, thrust impulses, and collision merges.

---

## ✨ Key Features

### 🌌 Precision 3D Gravitational Physics
- **Symplectic Velocity Verlet Integrator**: Second-order time-reversible integration preserving total energy and angular momentum over long-term planetary epochs.
- **Sub-Stepped Physics Simulation**: Multi-substepping (default: 5 sub-steps per frame) eliminates numerical drift during close high-speed perihelion passages.
- **Plummer Softening Factor ($\epsilon$)**: Avoids artificial hyper-velocity slingshot singularities when point masses experience near-zero distance encounters.
- **Configurable Collision Physics**:
  - **Inelastic Merging**: Realistic planetary coalescence adhering strictly to conservation of linear momentum ($m_1 \vec{v}_1 + m_2 \vec{v}_2 = M \vec{v}$) and spherical volume scaling ($R = \sqrt[3]{R_1^3 + R_2^3}$).
  - **Elastic Bouncing**: Restitution-based impulse deflection with contact normal separation to model dense asteroid fields.
  - **Ghost Mode**: Non-colliding pass-through masses for complex theoretical trajectory studies.

### ⚓ Stationary Mass Anchors
- Toggle any body into a **Fixed Gravitational Anchor** (`IsStationary`).
- Create immovable central stars, static black holes, artificial gravitational wells, or anchor points without computational runaway.

### 🚀 Interactive Orbital Maneuvering & Flight Computer
- **Raycast 3D Selection**: Click directly on any planet, moon, or star in perspective space to select it.
- **Auto-Circularize Orbit**: Instantaneous calculation and application of circular orbital velocity ($v_c = \sqrt{GM/r}$) relative to the primary gravitational parent.
- **Prograde & Retrograde Thrust**: Accelerate (+10%) or decelerate (-10%) along the current instantaneous velocity vector to elevate apastron or de-orbit into atmospheric capture.
- **Inclination Boosts (+Y / -Y)**: Apply out-of-plane orbital plane changes to tilt trajectories into high 3D inclinations.
- **Live Mass & Radius Scaling**: Dynamically scale mass (-50%, -10%, +10%, x2) or alter radius in real time to observe instantaneous orbital perturbations.

### 🎯 Slingshot Spawner & Orbital Sandbox
- Switch to **Spawner Mode** to place custom celestial bodies directly onto the equatorial reference plane.
- **Interactive Drag Slingshot**: Click and drag in 3D to establish position and fling new masses into custom trajectories with real-time vector previews.
- Multiple celestial presets available for spawning: **Terrestrial Planets, Moons / Asteroids, Gas Giants, Protostars, and Micro Black Holes**.

### 🎥 6-DOF Orbital Camera & Body Tracking
- Intuitive **Orbit**, **Pan**, and **Smooth Zoom** camera controls.
- **Body Tracking Mode (`C`)**: Smooth camera auto-tracking that follows any moving planet, moon, or rogue star through the cosmos.
- **Instant Focus (`F`)**: Quick-center view directly onto the selected object.

### 📊 Visual Telemetry & HUD
- **Persistent Orbital Trails**: Fading color-coded ribbon paths rendering past trajectories.
- **Force & Velocity Vectors**: Toggle real-time velocity (green) and resultant gravitational pull (orange) vectors.
- **Screen-Space Nameplates**: Projected 2D nameplates displaying mass, speed, and status over each body.
- **Altitude Drop-Lines**: Projection lines connecting 3D bodies to the equatorial reference grid.
- **Atmospheric Glow & Starfield**: Emissive star coronas and twinkling 3D cosmic background stars.

---

## 📸 Screenshot Gallery

<div align="center">

### ☀️ The Solar System
![Solar System Simulation](assets/screenshots/solar_system.png)
*Sun, inner terrestrial planets, Earth-Moon hierarchical orbit, Jupiter with Galilean moons, Saturn's 16-particle rings, and Comet Halley.*

---

### ⭐⭐ Binary Star Choreography
![Binary Star System](assets/screenshots/binary_stars.png)
*Two massive stars in mutual orbit around their common barycenter, orbited by stable circumbinary planets.*

---

### 🌀 Chaotic 3-Body Choreography
![Three-Body Problem](assets/screenshots/three_body.png)
*The legendary 3-Body Problem: three stars in chaotic gravitational interaction with trapped rogue planets weaving complex ribbons.*

---

### 🌌 Supermassive Accretion Disk
![Galaxy Disk](assets/screenshots/galaxy_disk.png)
*A central supermassive singularity holding 60 swirling Keplerian dust and stellar masses in a dense orbital accretion disk.*

---

### 🛠️ Body Inspector & Orbital Maneuvers
![Inspector Panel & Thrust](assets/screenshots/inspector_panel.png)
*Interactive Inspector panel: lock stationary state, tune mass, apply prograde/retrograde delta-v burns, and circularize orbits.*

---

### 🎯 Slingshot Spawner
![Interactive Slingshot Spawner](assets/screenshots/spawner_slingshot.png)
*Click-and-drag 3D slingshot vector system: launch custom planets, moons, or black holes directly into orbit.*

</div>

---

## 🪐 Preset Scenarios

| Preset | Description | Key Highlights |
| :--- | :--- | :--- |
| **1. Solar System** | Full planetary simulation from Mercury out to Neptune | Earth-Moon hierarchical orbit, Jupiter with Io & Europa, Saturn with 16 ring particles, and inclined eccentric Comet Halley |
| **2. Binary Stars** | Twin co-orbiting stars of equal mass | Balanced barycenter rotation with 4 distant circumbinary planets (Tatooine I & II, Oricon, Arrakis) |
| **3. 3-Body Problem** | Classic chaotic gravitational choreography | Three equal stars in triangular orbital resonance with trapped rogue planetoids displaying unpredictable trajectories |
| **4. Galaxy Disk** | Relativistic accretion disk around a singularity | 60 dust and stellar masses revolving in Keplerian velocity gradients around a central supermassive black hole |
| **5. Empty Sandbox** | Blank cosmic void | Clean universe ready for custom multi-body architectures, slingshot launches, and gravitational experiments |

---

## 🎮 Gameplay & Flight Computer

```
                           [ TOP CONTROL BAR ]
  +-------------------------------------------------------------------------+
  | GRAVITY SIM  [PAUSE] [<] 1.0x [>] | [Solar] [Binary] [3-Body] [Disk] [Reset] |
  +-------------------------------------------------------------------------+
                                      |
                     3D VIEWPORT      |---> [ INSPECTOR PANEL ] (Right)
               (Orbit / Pan / Zoom)   |     - Body Name & Fixed Lock Toggle
                                      |     - Mass Scaling: [-50%] [-10%] [+10%] [x2]
                                      |     - Radius Adjusters [-] [+]
                                      |     - Velocity & Speed Telemetry
                                      |     - [Auto Circularize Orbit (Auto-V)]
                                      |     - [Prograde +10%]  [Retrograde -10%]
                                      |     - [Inclination +Y] [Inclination -Y]
                                      |     - [Zero Velocity (Stop Dead)]
                                      |     - [Track With Camera: ON/OFF]
                                      |     - [DELETE BODY]
                                      |
  +-------------------------------------------------------------------------+
  | [SPAWN: ON/OFF] | [Trails] [Grid] [Vectors] [Forces] [Labels] [Collision]   |
  +-------------------------------------------------------------------------+
                           [ BOTTOM TOOLBAR ]
```

### Performing Orbital Maneuvers
1. **Left-Click** any celestial body in the 3D space to open its **Inspector Panel**.
2. **Auto-Circularize**: Click `Circularize Orbit (Auto-V)` to instantaneously compute the Keplerian orbital speed relative to the dominant gravitational body ($v = \sqrt{GM/r}$) and inject the body into a stable, closed circular orbit.
3. **Change Orbital Altitude**:
   - Click `Prograde (+10%)` to raise the apoapsis on the opposite side of the orbit.
   - Click `Retrograde (-10%)` to lower the periapsis and tighten the orbit.
4. **Change Orbital Plane**: Click `Inclination +Y` or `Inclination -Y` to induce 3D orbital tilt.
5. **Fixed Gravity Well**: Click `Make Stationary (Lock)` to anchor the body at its current position, turning it into a fixed gravitating center.

### Launching Bodies with the Slingshot Spawner
1. Toggle the **SPAWN** button on the bottom toolbar.
2. Select your body archetype: **Earth**, **Moon**, **Giant**, **Star**, or **BlackHole**.
3. **Left-Click & Drag** anywhere on the reference grid:
   - The origin point indicates the spawn coordinates.
   - The yellow trajectory vector defines the initial velocity magnitude and launch direction.
4. Release the mouse button to fling the new celestial mass into orbit!

---

## ⌨️ Controls & Hotkeys

### Mouse Controls
| Input | Action |
| :--- | :--- |
| **Right Mouse Drag** | Orbit / Rotate 3D camera around target |
| **Middle Mouse Drag** *(or `Shift` + Right Drag)* | Pan camera plane |
| **Mouse Wheel** | Zoom in / Zoom out |
| **Left Click (3D Object)** | Select body & open Inspector Panel |
| **Left Click (Empty Space)** | Deselect current body |
| **Left Click + Drag (Spawn Mode)** | Aim & launch slingshot body |

### Keyboard Shortcuts
| Key | Action |
| :--- | :--- |
| <kbd>Space</kbd> | Pause / Resume physics simulation |
| <kbd>1</kbd> - <kbd>5</kbd> | Instant preset selector (Solar System, Binary, 3-Body, Galaxy, Empty) |
| <kbd>W</kbd> / <kbd>A</kbd> / <kbd>S</kbd> / <kbd>D</kbd> *(or Arrow Keys)* | Pan camera forward / left / backward / right |
| <kbd>Q</kbd> / <kbd>E</kbd> | Elevate camera down / up |
| <kbd>F</kbd> | Focus camera on selected body |
| <kbd>C</kbd> | Toggle camera auto-tracking selected body |
| <kbd>R</kbd> | Reset camera to default orbital viewpoint |
| <kbd>Del</kbd> / <kbd>Backspace</kbd> | Destroy selected body |
| <kbd>F1</kbd> | Toggle in-game help manual overlay |
| <kbd>F12</kbd> | Take high-resolution screenshot (saved to `screenshots/`) |

---

## 🖥️ Supported Platforms

Gravity Mass Simulator is built on pure Go and [raylib-go](https://github.com/gen2brain/raylib-go), compiled with Cgo, and fully tested across modern desktop platforms:

| Platform | Architecture | Graphics Backend | Status |
| :--- | :--- | :--- | :---: |
| 🐧 **Linux** | `x86_64`, `arm64`, `riscv64` | OpenGL 3.3+ (X11 / Wayland) | ✅ Fully Supported |
| 🪟 **Windows** | `x86_64`, `ARM64` | OpenGL 3.3+ (DirectX / WGL) | ✅ Fully Supported |
| 🍎 **macOS** | `Apple Silicon (M1/M2/M3/M4)`, `Intel x86_64` | OpenGL 3.3 Core (macOS 10.15+) | ✅ Fully Supported |
| 🎮 **Steam Deck / Handhelds** | `x86_64` | Proton / Native Linux OpenGL | ✅ Tested & Playable |

### System Requirements
- **OS**: Linux (glibc 2.27+), Windows 10/11 (64-bit), macOS 11.0+
- **Processor**: Dual-Core 2.0 GHz or higher
- **Memory**: 2 GB RAM
- **Graphics**: GPU with OpenGL 3.3 Core Profile support (Intel HD 4000+, NVIDIA GeForce 400+, AMD Radeon HD 5000+)
- **Storage**: ~50 MB available space

---

## 🚀 Installation & Quick Start

### Prerequisites
Make sure you have **[Go](https://go.dev/dl/)** (1.21 or later) and a standard C compiler installed.

#### Linux (Debian / Ubuntu / Pop!_OS)
```bash
sudo apt update
sudo apt install -y build-essential libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev
```

#### Linux (Arch / Manjaro)
```bash
sudo pacman -S --needed base-devel mesa libxi libxcursor libxrandr libxinerama wayland libxkbcommon
```

#### Linux (Fedora / RHEL)
```bash
sudo dnf install -y gcc make mesa-libGL-devel libXi-devel libXcursor-devel libXrandr-devel libXinerama-devel wayland-devel libxkbcommon-devel
```

#### macOS
```bash
xcode-select --install
```

#### Windows
Install **[Go](https://go.dev/dl/)** and **[w64devkit](https://github.com/skeeto/w64devkit)** or MSYS2 MinGW-w64.

---

### 📦 Build and Run

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-username/gravitysim.git
   cd gravitysim
   ```

2. **Download Go module dependencies**:
   ```bash
   go mod download
   ```

3. **Run directly**:
   ```bash
   go run .
   ```

4. **Or compile an optimized standalone binary**:
   ```bash
   # Build executable
   go build -ldflags="-s -w" -o gravitysim .

   # Launch simulator
   ./gravitysim
   ```

---

## 🧮 Mathematical & Physics Engine

### 1. Newton's Universal Gravitation with Softening
For any body $i$ influenced by all other masses $j \neq i$:

$$\vec{F}_i = \sum_{j \neq i} G \frac{m_i m_j}{\left( \|\vec{r}_j - \vec{r}_i\|^2 + \epsilon^2 \right)^{3/2}} (\vec{r}_j - \vec{r}_i)$$

Where:
- $G$ = Universal gravitational constant (dynamically adjustable).
- $\epsilon$ = Plummer softening radius to eliminate infinite acceleration singularities during near-field interactions.

### 2. Symplectic Velocity Verlet Integration
Unlike Euler or standard Runge-Kutta, the Velocity Verlet integrator is **symplectic**, conserving the Hamiltonian phase space area and bounding long-term orbital energy error:

$$\vec{x}(t + \Delta t) = \vec{x}(t) + \vec{v}(t)\Delta t + \frac{1}{2}\vec{a}(t)\Delta t^2$$

$$\vec{v}(t + \Delta t) = \vec{v}(t) + \frac{1}{2}\left[\vec{a}(t) + \vec{a}(t + \Delta t)\right]\Delta t$$

### 3. Inelastic Momentum & Volume Conservation
When two celestial bodies collide in `Merge` mode:

$$\vec{v}_{\text{merged}} = \frac{m_1 \vec{v}_1 + m_2 \vec{v}_2}{m_1 + m_2}$$

$$R_{\text{merged}} = \left( R_1^3 + R_2^3 \right)^{1/3}$$

---

## 🗺️ Project Architecture

```
gravitysim/
├── assets/                  # Visual showcases and promotional assets
│   ├── hero.png             # Promotional banner
│   ├── showcase_preview.gif # High-framerate animated preview
│   ├── showcase_web.mp4     # Web-optimized H.264 video showcase
│   ├── showcase.mp4         # Master showcase video
│   └── screenshots/         # In-game scenario captures
│       ├── solar_system.png
│       ├── binary_stars.png
│       ├── three_body.png
│       ├── galaxy_disk.png
│       ├── inspector_panel.png
│       └── spawner_slingshot.png
├── camera.go                # 6-DOF orbital camera, zoom, pan, and body tracking
├── demo_capture.go          # Automated promotional capture and ffmpeg encoder
├── main.go                  # Main entry point, event pump, and hotkeys
├── physics.go               # Velocity Verlet integrator, N-body gravity, collisions
├── presets.go               # Solar system, binary stars, 3-body, and galaxy scenarios
├── render.go                # 3D viewport rendering, trails, starfield, vectors
├── types.go                 # Data structures, config, and state definitions
├── ui.go                    # 2D HUD, inspector panel, toolbars, and help manual
├── go.mod                   # Go module definition
└── README.md                # Project documentation
```

---

## 🔮 Roadmap & Upcoming Features

- [ ] **Barnes-Hut Octree ($O(N \log N)$)**: Spatial tree partitioning to scale simulations to $10,000+$ stars.
- [ ] **Relativistic Post-Newtonian Precession**: General relativity corrections modeling perihelion advance around black holes.
- [ ] **Scenario JSON Import / Export**: Save and share custom universe architectures.
- [ ] **Custom Textured Spheres & Normal Maps**: High-detail celestial surface textures.
- [ ] **Roche Limit Tidal Disruption**: Gravitational tidal tearing of moons straying within planetary Roche thresholds.

---

## 🤝 Contributing

Contributions, feature requests, and bug reports are welcome!
1. Fork the Project.
2. Create your Feature Branch (`git checkout -b feature/CosmicFeature`).
3. Commit your Changes (`git commit -m 'Add new cosmic feature'`).
4. Push to the Branch (`git push origin feature/CosmicFeature`).
5. Open a Pull Request.

---

## 📄 License

Distributed under the **MIT License**. See `LICENSE` for more information.

---

<div align="center">
  <sub>Engineered with 🪐 gravity and precision in Go. Star ⭐ this repository if you enjoyed exploring the cosmos!</sub>
</div>
