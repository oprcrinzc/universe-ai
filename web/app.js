// Gravity Mass Simulator 3D - WebAssembly Edition
// Full 3D WebGL N-Body Simulation Engine

let scene, camera, renderer, controls;
let bodiesMap = new Map(); // id -> THREE.Mesh
let trailsMap = new Map(); // id -> THREE.Line
let spacetimeMesh = null;
let vectorFieldLines = null;
let starfieldPoints = null;
let raycaster, mouse;
let selectedBodyId = -1;
let isSpawning = false;
let spawnStartPos = null;
let spawnAimLine = null;
let lastFrameTime = performance.now();
let fpsCounter = 0, fpsTimer = 0, currentFps = 144;
let isSimulationPaused = false;
let simSpeed = 1.0;
let timeScaleStep = 0.5;

let toastTimer = null;
function showToast(msg) {
    const toast = document.getElementById('toast');
    if (!toast) return;
    toast.innerText = msg;
    toast.classList.remove('hidden');
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        toast.classList.add('hidden');
    }, 1500);
}

function setSimulationSpeed(newSpeed) {
    newSpeed = Math.round(newSpeed * 10) / 10;
    if (newSpeed < 0.1) newSpeed = 0.1;
    if (newSpeed > 10.0) newSpeed = 10.0;
    simSpeed = newSpeed;

    const sel = document.getElementById('select-speed');
    if (sel) {
        let matched = false;
        for (let opt of sel.options) {
            if (Math.abs(parseFloat(opt.value) - simSpeed) < 0.05) {
                sel.value = opt.value;
                matched = true;
                break;
            }
        }
        if (!matched) {
            const opt = document.createElement('option');
            opt.value = simSpeed.toFixed(1);
            opt.innerText = simSpeed.toFixed(1) + 'x';
            sel.appendChild(opt);
            sel.value = opt.value;
        }
    }
}

function adjustSimulationSpeed(delta) {
    const newSpeed = Math.round((simSpeed + delta) * 10) / 10;
    const clamped = Math.max(0.1, Math.min(10.0, newSpeed));
    setSimulationSpeed(clamped);
    const sign = delta > 0 ? '+' : '';
    showToast(`Speed: ${clamped.toFixed(1)}x (${sign}${delta.toFixed(1)}x)`);
}

// Config toggles
let showTrails = true;
let showGrid = true;
let showSpacetime = false;
let showVectorField = false;
let showVelocityGlow = false;
let spacetimeScale = 1.0;
const SPACETIME_STEPS = 120;
const STATIC_SPACETIME_SPAN = 300.0;
const VECTOR_FIELD_STEPS = 32;

// 1. Initialize WebAssembly
async function initWasm() {
    try {
        const go = new Go();
        const response = await fetch('gravitysim.wasm');
        if (!response.ok) {
            throw new Error(`Failed to fetch gravitysim.wasm: ${response.statusText}`);
        }
        const wasmBytes = await response.arrayBuffer();
        const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
        go.run(result.instance);

        console.log("[WASM] GravitySim WebAssembly Engine Initialized Successfully!");
        document.getElementById('loading-overlay').style.opacity = '0';
        setTimeout(() => document.getElementById('loading-overlay').remove(), 500);

        initThreeJS();
        setupEventListeners();
        animate();
    } catch (err) {
        console.error("[WASM Error]", err);
        document.querySelector('.loading-box h2').innerText = "Initialization Error";
        document.querySelector('.loading-box p').innerText = err.message;
    }
}

// 2. Setup Three.js WebGL Renderer
function initThreeJS() {
    const container = document.getElementById('canvas-container');
    const w = container.clientWidth;
    const h = container.clientHeight;

    scene = new THREE.Scene();
    scene.background = new THREE.Color(0x06080e);

    camera = new THREE.PerspectiveCamera(55, w / h, 0.5, 5000);
    camera.position.set(0, 75, 130);

    renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: "high-performance" });
    renderer.setSize(w, h);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    container.appendChild(renderer.domElement);

    controls = new THREE.OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.maxDistance = 2500;
    controls.minDistance = 2;

    // Ambient & Directional Lights
    const ambientLight = new THREE.AmbientLight(0x334466, 1.2);
    scene.add(ambientLight);

    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight.position.set(50, 100, 50);
    scene.add(dirLight);

    // Coordinate Reference Grid
    const gridHelper = new THREE.GridHelper(300, 30, 0x1f2e4d, 0x141d30);
    gridHelper.position.y = -0.1;
    gridHelper.name = "gridHelper";
    scene.add(gridHelper);

    // Deep Starfield Background
    createStarfield();

    // Spacetime Curvature Grid Mesh
    createSpacetimeGridMesh();

    // Slingshot aim line
    const aimMat = new THREE.LineBasicMaterial({ color: 0x00d2ff, linewidth: 2 });
    const aimGeo = new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(), new THREE.Vector3()]);
    spawnAimLine = new THREE.Line(aimGeo, aimMat);
    spawnAimLine.visible = false;
    scene.add(spawnAimLine);

    raycaster = new THREE.Raycaster();
    mouse = new THREE.Vector2();

    window.addEventListener('resize', onWindowResize);
}

function createStarfield() {
    const starCount = 2000;
    const geom = new THREE.BufferGeometry();
    const positions = new Float32Array(starCount * 3);
    const colors = new Float32Array(starCount * 3);

    for (let i = 0; i < starCount; i++) {
        const r = 800 + Math.random() * 800;
        const theta = Math.random() * Math.PI * 2;
        const phi = Math.acos(2 * Math.random() - 1);
        positions[i * 3 + 0] = r * Math.sin(phi) * Math.cos(theta);
        positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
        positions[i * 3 + 2] = r * Math.cos(phi);

        const lum = 0.6 + Math.random() * 0.4;
        colors[i * 3 + 0] = lum;
        colors[i * 3 + 1] = lum;
        colors[i * 3 + 2] = lum + Math.random() * 0.2;
    }
    geom.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geom.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    const mat = new THREE.PointsMaterial({ size: 1.5, vertexColors: true });
    starfieldPoints = new THREE.Points(geom, mat);
    scene.add(starfieldPoints);
}

function getStaticSpacetimeSpan() {
    return STATIC_SPACETIME_SPAN * spacetimeScale;
}

function createSpacetimeGridMesh() {
    if (spacetimeMesh) {
        scene.remove(spacetimeMesh);
        spacetimeMesh.geometry.dispose();
        spacetimeMesh.material.dispose();
    }
    const span = getStaticSpacetimeSpan();
    const steps = SPACETIME_STEPS;
    const geom = new THREE.PlaneGeometry(span, span, steps, steps);
    geom.rotateX(-Math.PI / 2);

    const count = (steps + 1) * (steps + 1);
    const colors = new Float32Array(count * 3);
    for (let i = 0; i < count; i++) {
        colors[i * 3 + 0] = 0.05;
        colors[i * 3 + 1] = 0.45;
        colors[i * 3 + 2] = 0.85;
    }
    geom.setAttribute('color', new THREE.BufferAttribute(colors, 3));

    const mat = new THREE.MeshBasicMaterial({
        vertexColors: true,
        wireframe: true,
        transparent: true,
        opacity: 0.40
    });
    spacetimeMesh = new THREE.Mesh(geom, mat);
    spacetimeMesh.position.set(0, 0, 0); // GLOBAL POSITION: fixed at world origin
    spacetimeMesh.visible = showSpacetime;
    spacetimeMesh.userData.currentSpan = span;
    scene.add(spacetimeMesh);
}

// 3. Main Animation Loop (Target 144 FPS)
function animate() {
    requestAnimationFrame(animate);

    const now = performance.now();
    let dt = (now - lastFrameTime) / 1000.0;
    lastFrameTime = now;
    if (dt > 0.05) dt = 0.05;

    // FPS Counter
    fpsCounter++;
    fpsTimer += dt;
    if (fpsTimer >= 0.5) {
        currentFps = Math.round(fpsCounter / fpsTimer);
        document.getElementById('hud-fps').innerText = currentFps;
        fpsCounter = 0;
        fpsTimer = 0;
    }

    // Step WebAssembly physics
    if (!isSimulationPaused && window.GravitySim) {
        const physTime = GravitySim.step(dt * simSpeed);
        document.getElementById('hud-phys-time').innerText = physTime.toFixed(2) + " ms";
    }

    // Update 3D Visuals
    updateCelestialBodies();
    updateSpacetimeGrid();
    updateVectorField();

    controls.update();
    renderer.render(scene, camera);

    // Update 2D Tactical Viewport & Heatmap
    if (show2DViewport) {
        render2DViewport();
    }

    // Update Inspector Telemetry
    if (selectedBodyId > 0) {
        updateInspectorData();
    }
}

// 4. Update Celestial Bodies from WebAssembly
function updateCelestialBodies() {
    if (!window.GravitySim) return;

    // Retrieve ultra-fast packed Float32Array
    const buffer = GravitySim.getBodiesBuffer();
    const stride = 16;
    const bodyCount = buffer.length / stride;
    document.getElementById('hud-bodies').innerText = bodyCount;

    const activeIds = new Set();

    for (let i = 0; i < bodyCount; i++) {
        const offset = i * stride;
        const id = Math.round(buffer[offset + 0]);
        const x = buffer[offset + 1];
        const y = buffer[offset + 2];
        const z = buffer[offset + 3];
        const vx = buffer[offset + 4];
        const vy = buffer[offset + 5];
        const vz = buffer[offset + 6];
        const radius = buffer[offset + 7];
        const mass = buffer[offset + 8];
        const cr = buffer[offset + 9];
        const cg = buffer[offset + 10];
        const cb = buffer[offset + 11];
        const flags = Math.round(buffer[offset + 13]);
        const isStar = (flags & 1) !== 0;
        const isBlackHole = (flags & 2) !== 0;
        const isPulsar = (flags & 4) !== 0;
        const isStationary = (flags & 16) !== 0;
        const isSelected = (flags & 32) !== 0;
        const speed = buffer[offset + 15];

        activeIds.add(id);

        let mesh = bodiesMap.get(id);
        if (!mesh) {
            // Create mesh
            const geo = new THREE.SphereGeometry(Math.max(0.2, radius), 16, 16);
            let col = new THREE.Color(cr, cg, cb);
            if (isBlackHole) col = new THREE.Color(0x05040a);

            let mat;
            if (isStar) {
                mat = new THREE.MeshBasicMaterial({ color: col });
            } else if (isBlackHole) {
                mat = new THREE.MeshBasicMaterial({ color: col });
            } else {
                mat = new THREE.MeshStandardMaterial({
                    color: col,
                    roughness: 0.65,
                    metalness: 0.1
                });
            }

            mesh = new THREE.Mesh(geo, mat);
            mesh.userData = { id, name: `Body #${id}`, isStar, isBlackHole, isPulsar };
            scene.add(mesh);
            bodiesMap.set(id, mesh);

            // Create trail if small swarm
            if (bodyCount < 150) {
                const trailMat = new THREE.LineBasicMaterial({
                    color: col,
                    transparent: true,
                    opacity: 0.35
                });
                const maxPts = 60;
                const trailGeo = new THREE.BufferGeometry();
                const posArr = new Float32Array(maxPts * 3);
                for (let k = 0; k < maxPts * 3; k += 3) {
                    posArr[k + 0] = x;
                    posArr[k + 1] = y;
                    posArr[k + 2] = z;
                }
                trailGeo.setAttribute('position', new THREE.BufferAttribute(posArr, 3));
                const trailLine = new THREE.Line(trailGeo, trailMat);
                trailLine.userData = { head: 0, count: 0, maxPts };
                scene.add(trailLine);
                trailsMap.set(id, trailLine);
            }
        }

        mesh.position.set(x, y, z);

        // Velocity glow mode
        if (showVelocityGlow && !isStar && !isBlackHole) {
            const glowRatio = Math.min(1.0, speed / 12.0);
            mesh.material.color.setHSL(0.58 - glowRatio * 0.55, 1.0, 0.5);
        }

        // Selection highlight
        if (isSelected) {
            mesh.scale.set(1.15, 1.15, 1.15);
        } else {
            mesh.scale.set(1.0, 1.0, 1.0);
        }

        // Update trail points
        if (showTrails && bodyCount < 150) {
            const trail = trailsMap.get(id);
            if (trail) {
                const posAttr = trail.geometry.attributes.position;
                const arr = posAttr.array;
                const head = trail.userData.head;
                arr[head * 3 + 0] = x;
                arr[head * 3 + 1] = y;
                arr[head * 3 + 2] = z;
                trail.userData.head = (head + 1) % trail.userData.maxPts;
                posAttr.needsUpdate = true;
            }
        }
    }

    // Remove defunct bodies
    for (const [id, mesh] of bodiesMap.entries()) {
        if (!activeIds.has(id)) {
            scene.remove(mesh);
            mesh.geometry.dispose();
            mesh.material.dispose();
            bodiesMap.delete(id);

            const trail = trailsMap.get(id);
            if (trail) {
                scene.remove(trail);
                trail.geometry.dispose();
                trail.material.dispose();
                trailsMap.delete(id);
            }
        }
    }
}

// 5. Update 3D Spacetime Grid
function updateSpacetimeGrid() {
    if (!showSpacetime || !window.GravitySim) return;

    const span = getStaticSpacetimeSpan();
    if (!spacetimeMesh || Math.abs(spacetimeMesh.userData.currentSpan - span) > 0.01) {
        createSpacetimeGridMesh();
    }

    const steps = SPACETIME_STEPS;
    const depths = GravitySim.getSpacetimeGrid(0, 0, span, steps);
    const posAttr = spacetimeMesh.geometry.attributes.position;
    const colorAttr = spacetimeMesh.geometry.attributes.color;
    const arr = posAttr.array;
    const cols = colorAttr.array;

    for (let i = 0; i < depths.length; i++) {
        // In PlaneGeometry rotated -PI/2, Y is height
        arr[i * 3 + 1] = depths[i];
        const depthRatio = Math.min(1.0, Math.abs(depths[i]) / 26.0);
        // Einstein potential well color gradient: deep space cyan -> teal -> gold -> fiery red
        cols[i * 3 + 0] = 0.05 + depthRatio * 0.95;
        cols[i * 3 + 1] = 0.45 + depthRatio * 0.35;
        cols[i * 3 + 2] = 0.85 - depthRatio * 0.65;
    }
    posAttr.needsUpdate = true;
    colorAttr.needsUpdate = true;
    spacetimeMesh.position.set(0, 0, 0); // Strictly global position
}

// 5b. Update Gravitational Vector Field (32x32 = 1,024 3D vectors)
function updateVectorField() {
    if (!showVectorField || !window.GravitySim) {
        if (vectorFieldLines) vectorFieldLines.visible = false;
        return;
    }

    const span = getStaticSpacetimeSpan();
    const steps = VECTOR_FIELD_STEPS;
    const buf = GravitySim.getVectorField(0, 0, span, steps);
    if (!buf || buf.length === 0) return;

    if (!vectorFieldLines || vectorFieldLines.userData.steps !== steps) {
        if (vectorFieldLines) {
            scene.remove(vectorFieldLines);
            vectorFieldLines.geometry.dispose();
            vectorFieldLines.material.dispose();
        }
        const geom = new THREE.BufferGeometry();
        const copyBuf = new Float32Array(buf.length);
        copyBuf.set(buf);
        const interBuf = new THREE.InterleavedBuffer(copyBuf, 6);
        geom.setAttribute('position', new THREE.InterleavedBufferAttribute(interBuf, 3, 0));
        geom.setAttribute('color', new THREE.InterleavedBufferAttribute(interBuf, 3, 3));
        const mat = new THREE.LineBasicMaterial({
            vertexColors: true,
            transparent: true,
            opacity: 0.85
        });
        vectorFieldLines = new THREE.LineSegments(geom, mat);
        vectorFieldLines.userData.steps = steps;
        vectorFieldLines.position.set(0, 0, 0); // Strictly global
        scene.add(vectorFieldLines);
    } else {
        const interBuf = vectorFieldLines.geometry.attributes.position.data;
        interBuf.array.set(buf);
        interBuf.needsUpdate = true;
    }
    vectorFieldLines.visible = true;
}

// 6. Inspector Telemetry
function updateInspectorData() {
    if (!window.GravitySim || selectedBodyId <= 0) return;

    const jsonStr = GravitySim.selectBody(selectedBodyId);
    if (!jsonStr) {
        document.getElementById('inspector-panel').classList.add('hidden');
        document.body.classList.remove('inspector-active');
        selectedBodyId = -1;
        return;
    }

    const data = JSON.parse(jsonStr);
    document.getElementById('inspector-panel').classList.remove('hidden');
    document.body.classList.add('inspector-active');
    document.getElementById('insp-name').innerText = data.name || `Body #${data.id}`;
    document.getElementById('insp-mass').innerText = data.mass.toFixed(2);
    document.getElementById('insp-radius').innerText = data.radius.toFixed(2);

    const spd = Math.sqrt(data.vel.x*data.vel.x + data.vel.y*data.vel.y + data.vel.z*data.vel.z);
    document.getElementById('insp-speed').innerText = spd.toFixed(2) + " km/s";

    if (data.elements && data.elements.primary_id > 0) {
        document.getElementById('insp-primary').innerText = data.elements.primary_name;
        document.getElementById('insp-dist').innerText = data.elements.distance.toFixed(1) + " AU";
        document.getElementById('insp-sma').innerText = data.elements.semi_major_axis.toFixed(1) + " AU";
        document.getElementById('insp-ecc').innerText = data.elements.eccentricity.toFixed(3);
        document.getElementById('insp-period').innerText = data.elements.period > 0 ? data.elements.period.toFixed(1) + " yrs" : "-";
        document.getElementById('insp-regime').innerText = data.elements.orbit_type;
    } else {
        document.getElementById('insp-primary').innerText = "None (Barycentric)";
        document.getElementById('insp-dist').innerText = "-";
        document.getElementById('insp-sma').innerText = "-";
        document.getElementById('insp-ecc').innerText = "-";
        document.getElementById('insp-period').innerText = "-";
        document.getElementById('insp-regime').innerText = "Unbound";
    }
}

// 7. Event Listeners & UI Controls
function setupEventListeners() {
    // Play / Pause
    const btnPlayPause = document.getElementById('btn-play-pause');
    btnPlayPause.addEventListener('click', () => {
        isSimulationPaused = !isSimulationPaused;
        btnPlayPause.innerText = isSimulationPaused ? '▶ RESUME' : '⏸ PAUSE';
        btnPlayPause.classList.toggle('primary', !isSimulationPaused);
    });

    // Speed Controls with constant stepping (±0.1x, ±0.5x, ±2.0x)
    const selectSpeed = document.getElementById('select-speed');
    if (selectSpeed) {
        selectSpeed.addEventListener('change', (e) => {
            setSimulationSpeed(parseFloat(e.target.value));
            showToast(`Speed: ${simSpeed.toFixed(1)}x`);
        });
    }

    const btnSpeedDown = document.getElementById('btn-speed-down');
    if (btnSpeedDown) {
        btnSpeedDown.addEventListener('click', () => {
            adjustSimulationSpeed(-timeScaleStep);
        });
        btnSpeedDown.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            const newSpeed = Math.max(0.1, Math.round((simSpeed * 0.5) * 10) / 10);
            setSimulationSpeed(newSpeed);
            showToast(`Speed: ${newSpeed.toFixed(1)}x (Halved)`);
        });
    }

    const btnSpeedUp = document.getElementById('btn-speed-up');
    if (btnSpeedUp) {
        btnSpeedUp.addEventListener('click', () => {
            adjustSimulationSpeed(timeScaleStep);
        });
        btnSpeedUp.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            const newSpeed = Math.min(10.0, Math.round((simSpeed * 2.0) * 10) / 10);
            setSimulationSpeed(newSpeed);
            showToast(`Speed: ${newSpeed.toFixed(1)}x (Doubled)`);
        });
    }

    const btnSpeedReset = document.getElementById('btn-speed-reset');
    if (btnSpeedReset) {
        btnSpeedReset.addEventListener('click', () => {
            setSimulationSpeed(1.0);
            showToast('Speed reset to 1.0x (Normal)');
        });
    }

    const btnSpeedStep = document.getElementById('btn-speed-step');
    if (btnSpeedStep) {
        btnSpeedStep.addEventListener('click', () => {
            if (timeScaleStep <= 0.15) {
                timeScaleStep = 0.5;
            } else if (timeScaleStep <= 0.6) {
                timeScaleStep = 2.0;
            } else {
                timeScaleStep = 0.1;
            }
            btnSpeedStep.innerText = `±${timeScaleStep.toFixed(1)}`;
            showToast(`Speed Step: ±${timeScaleStep.toFixed(1)}x`);
        });
    }

    // Preset Loader
    document.getElementById('btn-load-preset').addEventListener('click', () => {
        const presetIdx = parseInt(document.getElementById('select-preset').value);
        clearSceneBodies();
        GravitySim.init(presetIdx);
        selectedBodyId = -1;
        document.getElementById('inspector-panel').classList.add('hidden');
        document.body.classList.remove('inspector-active');
    });

    // Lock to Nearest Object
    function lockToNearestObject() {
        if (!bodiesMap || bodiesMap.size === 0) return;
        let nearestMesh = null;
        let nearestId = -1;
        let minDist = Infinity;

        const skipId = (selectedBodyId > 0 && bodiesMap.size > 1) ? selectedBodyId : -1;

        for (let [id, mesh] of bodiesMap.entries()) {
            if (id === skipId) continue;
            const d = camera.position.distanceTo(mesh.position);
            if (d < minDist) {
                minDist = d;
                nearestMesh = mesh;
                nearestId = id;
            }
        }

        if (!nearestMesh && skipId !== -1) {
            nearestMesh = bodiesMap.get(skipId);
            nearestId = skipId;
        }

        if (nearestMesh && nearestId > 0) {
            selectedBodyId = nearestId;
            controls.target.copy(nearestMesh.position);
            controls.update();
            if (window.GravitySim) GravitySim.selectBody(nearestId);
            updateInspectorData();
        }
    }

    const btnLockNear = document.getElementById('btn-lock-near');
    if (btnLockNear) {
        btnLockNear.addEventListener('click', lockToNearestObject);
    }

    // Reset Camera
    document.getElementById('btn-reset-cam').addEventListener('click', () => {
        camera.position.set(0, 75, 130);
        controls.target.set(0, 0, 0);
        controls.update();
    });

    // Help Modal
    const helpModal = document.getElementById('help-modal');
    document.getElementById('btn-help').addEventListener('click', () => helpModal.classList.remove('hidden'));
    document.getElementById('btn-close-help').addEventListener('click', () => helpModal.classList.add('hidden'));

    // Inspector Close
    document.getElementById('btn-close-insp').addEventListener('click', () => {
        document.getElementById('inspector-panel').classList.add('hidden');
        document.body.classList.remove('inspector-active');
        selectedBodyId = -1;
        GravitySim.selectBody(-1);
    });

    // Inspector Thrust Maneuvers
    document.getElementById('btn-circ').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "circularize", 1.0));
    document.getElementById('btn-prograde').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "prograde", 0.1));
    document.getElementById('btn-retrograde').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "retrograde", 0.1));
    document.getElementById('btn-stop').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "stop", 0));
    document.getElementById('btn-lock').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "toggle_stationary", 0));
    document.getElementById('btn-double-mass').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "scale_mass", 2.0));
    document.getElementById('btn-half-mass').addEventListener('click', () => GravitySim.applyThrust(selectedBodyId, "scale_mass", 0.5));
    document.getElementById('btn-delete').addEventListener('click', () => {
        GravitySim.deleteBody(selectedBodyId);
        document.getElementById('inspector-panel').classList.add('hidden');
        document.body.classList.remove('inspector-active');
        selectedBodyId = -1;
    });

    // Spawner Toggle
    const btnSpawn = document.getElementById('btn-spawn-mode');
    btnSpawn.addEventListener('click', () => {
        isSpawning = !isSpawning;
        btnSpawn.innerText = isSpawning ? 'SPAWN: ON' : 'SPAWN: OFF';
        btnSpawn.classList.toggle('active', isSpawning);
    });

    // Visualizer Toggles
    const togTrails = document.getElementById('tog-trails');
    togTrails.addEventListener('click', () => {
        showTrails = !showTrails;
        togTrails.classList.toggle('active', showTrails);
        for (const trail of trailsMap.values()) {
            trail.visible = showTrails;
        }
    });

    const togGrid = document.getElementById('tog-grid');
    togGrid.addEventListener('click', () => {
        showGrid = !showGrid;
        togGrid.classList.toggle('active', showGrid);
        scene.getObjectByName('gridHelper').visible = showGrid;
    });

    const togSpacetime = document.getElementById('tog-spacetime');
    togSpacetime.addEventListener('click', () => {
        showSpacetime = !showSpacetime;
        togSpacetime.classList.toggle('active', showSpacetime);
        spacetimeMesh.visible = showSpacetime;
    });

    const togVecField = document.getElementById('tog-vecfield');
    if (togVecField) {
        togVecField.addEventListener('click', () => {
            showVectorField = !showVectorField;
            togVecField.classList.toggle('active', showVectorField);
            if (vectorFieldLines) vectorFieldLines.visible = showVectorField;
        });
    }

    const togGlow = document.getElementById('tog-glow');
    togGlow.addEventListener('click', () => {
        showVelocityGlow = !showVelocityGlow;
        togGlow.classList.toggle('active', showVelocityGlow);
    });

    // 2D Labels Toggle (K)
    const tog2DLabels = document.getElementById('tog-2dlabels');
    if (tog2DLabels) {
        tog2DLabels.addEventListener('click', () => {
            show2DLabels = !show2DLabels;
            tog2DLabels.classList.toggle('active', show2DLabels);
            showToast(`2D Map Labels: ${show2DLabels ? 'ON' : 'OFF (Clean Mode)'}`);
        });
    }

    // 3D Heatmap Plane Toggle
    const tog3DHeatmap = document.getElementById('tog-3dheatmap');
    if (tog3DHeatmap) {
        tog3DHeatmap.addEventListener('click', () => {
            show3DHeatmapPlane = !show3DHeatmapPlane;
            tog3DHeatmap.classList.toggle('active', show3DHeatmapPlane);
            showToast(`3D Heatmap Plane: ${show3DHeatmapPlane ? 'ON' : 'OFF'}`);
        });
    }

    // Lagrange Points Toggle
    const togLagrange = document.getElementById('tog-lagrange');
    if (togLagrange) {
        togLagrange.addEventListener('click', () => {
            showLagrangePoints = !showLagrangePoints;
            togLagrange.classList.toggle('active', showLagrangePoints);
            showToast(`Lagrange Points L1-L5: ${showLagrangePoints ? 'ON' : 'OFF'}`);
        });
    }

    // Physics Engine Toggles
    const togBarnesHut = document.getElementById('tog-barnes-hut');
    togBarnesHut.addEventListener('click', () => {
        const active = togBarnesHut.classList.toggle('active');
        GravitySim.setConfig("enable_barnes_hut", active);
        document.getElementById('hud-engine').innerText = active ? "Barnes-Hut O(N log N)" : "Direct O(N^2) Exact";
    });

    const togRelativity = document.getElementById('tog-relativity');
    togRelativity.addEventListener('click', () => {
        const active = togRelativity.classList.toggle('active');
        GravitySim.setConfig("enable_relativity", active);
    });

    const togRoche = document.getElementById('tog-roche');
    togRoche.addEventListener('click', () => {
        const active = togRoche.classList.toggle('active');
        GravitySim.setConfig("enable_roche_limit", active);
    });

    const selectCollision = document.getElementById('select-collision');
    selectCollision.addEventListener('change', (e) => {
        GravitySim.setConfig("collision", parseInt(e.target.value));
    });

    // Click to Inspect / Slingshot Spawner
    const container = document.getElementById('canvas-container');
    container.addEventListener('pointerdown', onPointerDown);
    container.addEventListener('pointermove', onPointerMove);
    container.addEventListener('pointerup', onPointerUp);

    // Keyboard Shortcuts
    window.addEventListener('keydown', (e) => {
        if (e.key === ' ') {
            e.preventDefault();
            btnPlayPause.click();
        } else if (e.key === 'b' || e.key === 'B') {
            togBarnesHut.click();
        } else if (e.key === 'p' || e.key === 'P') {
            togSpacetime.click();
        } else if (e.key === 'o' || e.key === 'O') {
            if (togVecField) togVecField.click();
        } else if (e.key === 'l' || e.key === 'L') {
            togGlow.click();
        } else if (e.key === 'g' || e.key === 'G') {
            togRelativity.click();
        } else if (e.key === 'r' || e.key === 'R') {
            document.getElementById('btn-reset-cam').click();
        } else if (e.key === 'n' || e.key === 'N') {
            const btn = document.getElementById('btn-lock-near');
            if (btn) btn.click();
        } else if (e.key === '[') {
            spacetimeScale = Math.max(0.5, spacetimeScale - 0.25);
            createSpacetimeGridMesh();
        } else if (e.key === ']') {
            spacetimeScale = Math.min(3.5, spacetimeScale + 0.25);
            createSpacetimeGridMesh();
        } else if (e.key === ',' || e.key === '<') {
            const step = e.shiftKey ? 0.1 : timeScaleStep;
            adjustSimulationSpeed(-step);
        } else if (e.key === '.' || e.key === '>') {
            const step = e.shiftKey ? 0.1 : timeScaleStep;
            adjustSimulationSpeed(step);
        } else if (e.key === '\\') {
            setSimulationSpeed(1.0);
            showToast('Speed reset to 1.0x (Normal)');
        } else if (e.key === 'm' || e.key === 'M') {
            const btn = document.getElementById('tog-2dmap');
            if (btn) btn.click();
        } else if (e.key === 'k' || e.key === 'K') {
            const btn = document.getElementById('tog-2dlabels');
            if (btn) btn.click();
        } else if (e.key === 'h' || e.key === 'H') {
            // Toggle both 2D heatmap and 3D heatmap plane together
            const btnH = document.getElementById('tog-heatmap');
            if (btnH) btnH.click();
            const btn3D = document.getElementById('tog-3dheatmap');
            if (btn3D && show2DHeatmap === show3DHeatmapPlane) {
                // sync state: only click 3D if they were in sync, otherwise user may want independent control
                // do nothing extra - H toggles 2D, user can click 3D Heatmap independently
            }
        } else if (e.key >= '0' && e.key <= '9') {
            const num = parseInt(e.key);
            const presetIdx = (num === 0) ? 9 : (num - 1);
            const presetSel = document.getElementById('select-preset');
            if (presetSel && presetSel.options[presetIdx]) {
                presetSel.selectedIndex = presetIdx;
                document.getElementById('btn-load-preset').click();
            }
        }
    });

    setup2DViewport();
}

function clearSceneBodies() {
    for (const mesh of bodiesMap.values()) {
        scene.remove(mesh);
        mesh.geometry.dispose();
        mesh.material.dispose();
    }
    bodiesMap.clear();

    for (const trail of trailsMap.values()) {
        scene.remove(trail);
        trail.geometry.dispose();
        trail.material.dispose();
    }
    trailsMap.clear();

    if (showSpacetime) {
        createSpacetimeGridMesh();
    }
}

function onPointerDown(e) {
    if (e.button !== 0) return; // Left click only
    const rect = renderer.domElement.getBoundingClientRect();
    mouse.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
    mouse.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;

    if (isSpawning) {
        // Slingshot initiation on XZ plane (Y=0)
        raycaster.setFromCamera(mouse, camera);
        const plane = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
        const hit = new THREE.Vector3();
        if (raycaster.ray.intersectPlane(plane, hit)) {
            spawnStartPos = hit.clone();
            spawnAimLine.visible = true;
            controls.enabled = false;
        }
    } else {
        // Body selection
        raycaster.setFromCamera(mouse, camera);
        const meshes = Array.from(bodiesMap.values());
        const intersects = raycaster.intersectObjects(meshes);
        if (intersects.length > 0) {
            const hitMesh = intersects[0].object;
            selectedBodyId = hitMesh.userData.id;
            updateInspectorData();
        }
    }
}

function onPointerMove(e) {
    if (isSpawning && spawnStartPos) {
        const rect = renderer.domElement.getBoundingClientRect();
        mouse.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
        mouse.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;
        raycaster.setFromCamera(mouse, camera);
        const plane = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
        const hit = new THREE.Vector3();
        if (raycaster.ray.intersectPlane(plane, hit)) {
            const pts = [spawnStartPos, hit];
            spawnAimLine.geometry.setFromPoints(pts);
        }
    }
}

function onPointerUp(e) {
    if (isSpawning && spawnStartPos) {
        const rect = renderer.domElement.getBoundingClientRect();
        mouse.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
        mouse.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;
        raycaster.setFromCamera(mouse, camera);
        const plane = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
        const hit = new THREE.Vector3();
        if (raycaster.ray.intersectPlane(plane, hit)) {
            // Drag vector: v = (start - hit) * scale
            const drag = spawnStartPos.clone().sub(hit).multiplyScalar(0.25);
            const spawnType = document.getElementById('select-spawn').value;

            if (spawnType === 'swarm50') {
                GravitySim.spawnCluster("cluster50", spawnStartPos.x, spawnStartPos.y, spawnStartPos.z, drag.x, drag.y, drag.z);
            } else if (spawnType === 'collapse100') {
                GravitySim.spawnCluster("collapse100", spawnStartPos.x, spawnStartPos.y, spawnStartPos.z, drag.x, drag.y, drag.z);
            } else if (spawnType === 'galaxy150') {
                GravitySim.spawnCluster("galaxy150", spawnStartPos.x, spawnStartPos.y, spawnStartPos.z, drag.x, drag.y, drag.z);
            } else {
                let name = "Custom Planet";
                let mass = 1.0, rad = 0.85, r = 100, g = 180, b = 255;
                let isStar = false, isBH = false;

                if (spawnType === 'mars') {
                    name = "Mars Planetoid"; mass = 0.2; rad = 0.55; r = 240; g = 100; b = 60;
                } else if (spawnType === 'jupiter') {
                    name = "Gas Giant"; mass = 25.0; rad = 2.2; r = 240; g = 200; b = 150;
                } else if (spawnType === 'star') {
                    name = "Solar Star"; mass = 800.0; rad = 3.5; r = 255; g = 220; b = 60; isStar = true;
                } else if (spawnType === 'blackhole') {
                    name = "Singularity"; mass = 2500.0; rad = 2.8; r = 15; g = 10; b = 25; isBH = true;
                }

                GravitySim.spawnBody(name,
                    spawnStartPos.x, spawnStartPos.y, spawnStartPos.z,
                    drag.x, drag.y, drag.z,
                    mass, rad, r, g, b, isStar, isBH);
            }
        }
        spawnStartPos = null;
        spawnAimLine.visible = false;
        controls.enabled = true;
    }
}

function onWindowResize() {
    const container = document.getElementById('canvas-container');
    const w = container.clientWidth;
    const h = container.clientHeight;
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
    renderer.setSize(w, h);
}

// ----------------------------------------------------
// 2D Tactical Viewport & Continuous Gravitational Heatmap
// ----------------------------------------------------
let show2DViewport = false;
let show2DLabels = true;
let show2DHeatmap = true;
let show3DHeatmapPlane = false;
let showLagrangePoints = true;
let viewport2DFullscreen = false;
let viewport2DZoom = 0.5;
let viewport2DPan = { x: 0, z: 0 };
let isDragging2D = false;
let lastMouse2D = { x: 0, y: 0 };

function setup2DViewport() {
    const container = document.getElementById('viewport-2d-container');
    const canvas = document.getElementById('canvas-2d');
    const tog2DMap = document.getElementById('tog-2dmap');
    const togHeatmap = document.getElementById('tog-heatmap');
    const btnLabels = document.getElementById('btn-2d-labels');
    const btnReset = document.getElementById('btn-2d-reset');
    const btnFullscreen = document.getElementById('btn-2d-fullscreen');
    const btnClose = document.getElementById('btn-close-2d');

    if (!container || !canvas) return;

    if (tog2DMap) {
        tog2DMap.addEventListener('click', () => {
            show2DViewport = !show2DViewport;
            tog2DMap.classList.toggle('active', show2DViewport);
            container.classList.toggle('hidden', !show2DViewport);
            if (show2DViewport) update2DCanvasSize();
        });
    }

    if (togHeatmap) {
        togHeatmap.addEventListener('click', () => {
            show2DHeatmap = !show2DHeatmap;
            togHeatmap.classList.toggle('active', show2DHeatmap);
            if (show2DHeatmap && !show2DViewport && tog2DMap) {
                show2DViewport = true;
                tog2DMap.classList.add('active');
                container.classList.remove('hidden');
                update2DCanvasSize();
            }
        });
    }

    if (btnLabels) {
        btnLabels.addEventListener('click', () => {
            show2DLabels = !show2DLabels;
            btnLabels.classList.toggle('active', show2DLabels);
            // sync bottom toolbar toggle
            const togL = document.getElementById('tog-2dlabels');
            if (togL) togL.classList.toggle('active', show2DLabels);
            showToast(`2D Labels: ${show2DLabels ? 'ON' : 'OFF (Clean Mode)'}`);
        });
    }

    if (btnReset) {
        btnReset.addEventListener('click', () => {
            viewport2DZoom = 0.5;
            viewport2DPan = { x: 0, z: 0 };
            const scaleEl = document.getElementById('viewport-2d-scale');
            if (scaleEl) scaleEl.innerText = `Zoom: 1.0x`;
        });
    }

    if (btnFullscreen) {
        btnFullscreen.addEventListener('click', () => {
            viewport2DFullscreen = !viewport2DFullscreen;
            container.classList.toggle('fullscreen', viewport2DFullscreen);
            update2DCanvasSize();
        });
    }

    if (btnClose) {
        btnClose.addEventListener('click', () => {
            show2DViewport = false;
            if (tog2DMap) tog2DMap.classList.remove('active');
            container.classList.add('hidden');
        });
    }

    // Panning & Zooming events on 2D canvas
    canvas.addEventListener('wheel', (e) => {
        e.preventDefault();
        const factor = e.deltaY < 0 ? 1.18 : 0.85;
        const rect = canvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;
        const cx = canvas.width * 0.5;
        const cy = canvas.height * 0.5;

        const worldX = viewport2DPan.x + (mouseX - cx) / viewport2DZoom;
        const worldZ = viewport2DPan.z + (mouseY - cy) / viewport2DZoom;

        viewport2DZoom = Math.max(0.005, Math.min(60.0, viewport2DZoom * factor));
        viewport2DPan.x = worldX - (mouseX - cx) / viewport2DZoom;
        viewport2DPan.z = worldZ - (mouseY - cy) / viewport2DZoom;

        const scaleEl = document.getElementById('viewport-2d-scale');
        if (scaleEl) {
            const zoomDisplay = (viewport2DZoom / 0.5).toFixed(2);
            scaleEl.innerText = `Zoom: ${zoomDisplay}x`;
        }
    });

    canvas.addEventListener('mousedown', (e) => {
        if (e.button === 0 || e.button === 2) {
            isDragging2D = true;
            lastMouse2D = { x: e.clientX, y: e.clientY };
        }
    });

    window.addEventListener('mousemove', (e) => {
        if (isDragging2D) {
            const dx = e.clientX - lastMouse2D.x;
            const dy = e.clientY - lastMouse2D.y;
            viewport2DPan.x -= dx / viewport2DZoom;
            viewport2DPan.z -= dy / viewport2DZoom;
            lastMouse2D = { x: e.clientX, y: e.clientY };
        }
    });

    window.addEventListener('mouseup', () => {
        isDragging2D = false;
    });

    canvas.addEventListener('contextmenu', (e) => e.preventDefault());
    window.addEventListener('resize', update2DCanvasSize);
}

function update2DCanvasSize() {
    const container = document.getElementById('viewport-2d-container');
    const canvas = document.getElementById('canvas-2d');
    if (!container || !canvas) return;
    if (viewport2DFullscreen) {
        canvas.width = container.clientWidth;
        canvas.height = container.clientHeight - 32;
    } else {
        canvas.width = 340;
        canvas.height = 260;
    }
}

function render2DViewport() {
    const canvas = document.getElementById('canvas-2d');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const w = canvas.width;
    const h = canvas.height;
    if (w <= 0 || h <= 0) return;

    // Background
    ctx.fillStyle = '#060913';
    ctx.fillRect(0, 0, w, h);

    const cx = w * 0.5;
    const cy = h * 0.5;
    const zoom = viewport2DZoom;
    const panX = viewport2DPan.x;
    const panZ = viewport2DPan.z;

    const toScreenX = (wx) => cx + (wx - panX) * zoom;
    const toScreenY = (wz) => cy + (wz - panZ) * zoom;
    const toWorldX = (sx) => panX + (sx - cx) / zoom;
    const toWorldZ = (sy) => panZ + (sy - cy) / zoom;

    // 1. Grid
    ctx.strokeStyle = 'rgba(30, 50, 80, 0.4)';
    ctx.lineWidth = 1;
    const gridSpacing = 50 * zoom;
    if (gridSpacing > 15) {
        const startX = ((cx - panX * zoom) % gridSpacing + gridSpacing) % gridSpacing;
        for (let x = startX; x < w; x += gridSpacing) {
            ctx.beginPath();
            ctx.moveTo(x, 0);
            ctx.lineTo(x, h);
            ctx.stroke();
        }
        const startY = ((cy - panZ * zoom) % gridSpacing + gridSpacing) % gridSpacing;
        for (let y = startY; y < h; y += gridSpacing) {
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(w, y);
            ctx.stroke();
        }
    }

    // Origin crosshair
    const ox = toScreenX(0);
    const oy = toScreenY(0);
    ctx.strokeStyle = 'rgba(70, 110, 180, 0.5)';
    ctx.beginPath();
    ctx.moveTo(ox, 0); ctx.lineTo(ox, h);
    ctx.moveTo(0, oy); ctx.lineTo(w, oy);
    ctx.stroke();

    if (!window.GravitySim) return;
    const buffer = GravitySim.getBodiesBuffer();
    const stride = 16;
    const bodyCount = buffer.length / stride;

    // 2. Gravitational Heatmap
    if (show2DHeatmap && bodyCount > 0) {
        const cellCols = 28;
        const cellRows = 20;
        const cellW = w / cellCols;
        const cellH = h / cellRows;

        for (let r = 0; r < cellRows; r++) {
            const sy = (r + 0.5) * cellH;
            const wz = toWorldZ(sy);
            for (let c = 0; c < cellCols; c++) {
                const sx = (c + 0.5) * cellW;
                const wx = toWorldX(sx);

                let fx = 0, fz = 0;
                const maxSample = Math.min(bodyCount, 60);
                for (let i = 0; i < maxSample; i++) {
                    const off = i * stride;
                    const bx = buffer[off + 1];
                    const bz = buffer[off + 3];
                    const bmass = buffer[off + 8];
                    const dx = bx - wx;
                    const dz = bz - wz;
                    const d2 = dx * dx + dz * dz + 1.0;
                    const d = Math.sqrt(d2);
                    const f = bmass / (d2 * d);
                    fx += dx * f;
                    fz += dz * f;
                }
                const fMag = Math.sqrt(fx * fx + fz * fz);
                const logVal = Math.log10(1.0 + fMag * 50.0);
                const t = Math.max(0, Math.min(1.0, logVal / 2.2));

                if (t > 0.05) {
                    let cr, cg, cb;
                    if (t < 0.25) {
                        const s = t / 0.25;
                        cr = Math.round(10 + 10 * s);
                        cg = Math.round(30 + 120 * s);
                        cb = Math.round(100 + 155 * s);
                    } else if (t < 0.50) {
                        const s = (t - 0.25) / 0.25;
                        cr = Math.round(20 + 20 * s);
                        cg = Math.round(150 + 80 * s);
                        cb = Math.round(255 - 150 * s);
                    } else if (t < 0.75) {
                        const s = (t - 0.50) / 0.25;
                        cr = Math.round(40 + 215 * s);
                        cg = Math.round(230 - 30 * s);
                        cb = Math.round(105 - 85 * s);
                    } else {
                        const s = (t - 0.75) / 0.25;
                        cr = 255;
                        cg = Math.round(200 - 150 * (1 - s));
                        cb = Math.round(20 + 235 * s);
                    }
                    ctx.fillStyle = `rgba(${cr},${cg},${cb},${(t * 0.45).toFixed(2)})`;
                    ctx.fillRect(c * cellW, r * cellH, cellW + 1, cellH + 1);
                }
            }
        }
    }

    // 3. Directional Vector Field Arrows
    if (show2DHeatmap && bodyCount > 0) {
        const vCols = 14;
        const vRows = 10;
        const vStepX = w / vCols;
        const vStepY = h / vRows;

        ctx.lineWidth = 1;
        for (let r = 0; r < vRows; r++) {
            const sy = (r + 0.5) * vStepY;
            const wz = toWorldZ(sy);
            for (let c = 0; c < vCols; c++) {
                const sx = (c + 0.5) * vStepX;
                const wx = toWorldX(sx);

                let fx = 0, fz = 0;
                const maxSample = Math.min(bodyCount, 40);
                for (let i = 0; i < maxSample; i++) {
                    const off = i * stride;
                    const bx = buffer[off + 1];
                    const bz = buffer[off + 3];
                    const bmass = buffer[off + 8];
                    const dx = bx - wx;
                    const dz = bz - wz;
                    const d2 = dx * dx + dz * dz + 1.0;
                    const d = Math.sqrt(d2);
                    const f = bmass / (d2 * d);
                    fx += dx * f;
                    fz += dz * f;
                }
                const fMag = Math.sqrt(fx * fx + fz * fz);
                if (fMag > 0.001) {
                    const arrowLen = Math.min(14, 4 + Math.log10(1.0 + fMag * 100) * 4);
                    const nx = fx / fMag;
                    const nz = fz / fMag;
                    ctx.strokeStyle = 'rgba(80, 200, 255, 0.45)';
                    ctx.beginPath();
                    ctx.moveTo(sx, sy);
                    ctx.lineTo(sx + nx * arrowLen, sy + nz * arrowLen);
                    ctx.stroke();
                }
            }
        }
    }

    // 4. Project Bodies
    ctx.font = '9px sans-serif';
    ctx.textAlign = 'center';
    for (let i = 0; i < bodyCount; i++) {
        const off = i * stride;
        const bx = buffer[off + 1];
        const bz = buffer[off + 3];
        const vx = buffer[off + 4];
        const vz = buffer[off + 6];
        const radius = buffer[off + 7];
        const mass = buffer[off + 8];
        const cr = Math.round(buffer[off + 9] * 255);
        const cg = Math.round(buffer[off + 10] * 255);
        const cb = Math.round(buffer[off + 11] * 255);
        const flags = Math.round(buffer[off + 13]);
        const isStar = (flags & 1) !== 0;

        const sx = toScreenX(bx);
        const sy = toScreenY(bz);

        if (sx < -40 || sx > w + 40 || sy < -40 || sy > h + 40) continue;

        let drawR = Math.max(2.0, Math.min(18.0, Math.log10(mass + 1.0) * 3.0 * zoom + 1.5));
        if (isStar) drawR = Math.max(4.0, drawR * 1.2);

        // Velocity vector indicator
        const vLen = Math.sqrt(vx * vx + vz * vz);
        if (vLen > 0.01) {
            const vScale = Math.min(25, vLen * 3.0 * zoom);
            ctx.strokeStyle = 'rgba(0, 255, 200, 0.6)';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(sx, sy);
            ctx.lineTo(sx + (vx / vLen) * vScale, sy + (vz / vLen) * vScale);
            ctx.stroke();
        }

        // Body circle
        ctx.fillStyle = `rgb(${cr},${cg},${cb})`;
        ctx.beginPath();
        ctx.arc(sx, sy, drawR, 0, Math.PI * 2);
        ctx.fill();

        if (isStar) {
            ctx.strokeStyle = 'rgba(255, 220, 100, 0.7)';
            ctx.lineWidth = 1.5;
            ctx.stroke();
        }

        // Label for heavy or selected bodies (controlled by show2DLabels toggle)
        if (show2DLabels && (mass >= 5.0 || (bodyCount < 30 && mass > 0.1))) {
            ctx.fillStyle = 'rgba(220, 235, 255, 0.85)';
            ctx.fillText(`M:${mass >= 100 ? mass.toFixed(0) : mass.toFixed(1)}`, sx, sy - drawR - 3);
        }
    }
}

// Start Initialization
window.addEventListener('DOMContentLoaded', initWasm);
