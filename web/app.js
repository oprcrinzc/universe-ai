// Gravity Mass Simulator 3D - WebAssembly Edition
// Full 3D WebGL N-Body Simulation Engine

let scene, camera, renderer, controls;
let bodiesMap = new Map(); // id -> THREE.Mesh
let trailsMap = new Map(); // id -> THREE.Line
let spacetimeMesh = null;
let vectorFieldLines = null;
let bodyVectorsLines = null;
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
let showGravitationalWaves = false;
let showVectorField = false;
let showVectors = false;
let showVelocityGlow = false;
let show2DCircles = true;
let spacetimeScale = 1.0;
let spacetimeResolution = 120;
const availableResolutions = [32, 48, 64, 80, 96, 120, 160, 200, 256];
let heatmap2DResolution = 64;
const availableHeatmapResolutions = [16, 24, 32, 48, 64, 80, 96, 128, 160, 200];

function setVectorsEnabled(enable) {
    showVectors = enable;
    const togVectors = document.getElementById('tog-vectors');
    const btn2dVec = document.getElementById('btn-2d-vec');
    if (togVectors) togVectors.classList.toggle('active', showVectors);
    if (btn2dVec) btn2dVec.classList.toggle('active', showVectors);
    if (bodyVectorsLines) bodyVectorsLines.visible = showVectors;
    if (window.GravitySim && window.GravitySim.setConfig) {
        GravitySim.setConfig("show_vectors", showVectors);
    }
    showToast(`Object Direction Vectors: ${showVectors ? 'ON' : 'OFF'}`);
}

function set2DCirclesEnabled(enable) {
    show2DCircles = enable;
    const togCircles = document.getElementById('tog-2dcircles');
    const btn2dCir = document.getElementById('btn-2d-cir');
    if (togCircles) togCircles.classList.toggle('active', show2DCircles);
    if (btn2dCir) btn2dCir.classList.toggle('active', show2DCircles);
    if (window.GravitySim && window.GravitySim.setConfig) {
        GravitySim.setConfig("show_2d_circles", show2DCircles);
    }
    showToast(`2D Circles & Dots: ${show2DCircles ? 'ON' : 'OFF'}`);
}

function setTrailsEnabled(enable) {
    showTrails = enable;
    const togTrails = document.getElementById('tog-trails');
    const btn2dTrl = document.getElementById('btn-2d-trl');
    if (togTrails) togTrails.classList.toggle('active', showTrails);
    if (btn2dTrl) btn2dTrl.classList.toggle('active', showTrails);
    for (const trail of trailsMap.values()) {
        trail.visible = showTrails;
    }
    showToast(`Orbit Reference Trails: ${showTrails ? 'ON' : 'OFF'}`);
}

function setResolution(res) {
    spacetimeResolution = res;
    if (window.GravitySim && window.GravitySim.setConfig) {
        GravitySim.setConfig("spacetime_resolution", res);
    }
    const lbl = document.getElementById('lbl-resolution');
    if (lbl) lbl.textContent = `${res}x${res}`;
    createSpacetimeGridMesh();
    showToast(`Spacetime Resolution: ${res}x${res} (${(res + 1) * (res + 1)} Vertices)`);
}

function decreaseResolution() {
    let target = availableResolutions[0];
    for (let i = availableResolutions.length - 1; i >= 0; i--) {
        if (availableResolutions[i] < spacetimeResolution) {
            target = availableResolutions[i];
            break;
        }
    }
    setResolution(target);
}

function increaseResolution() {
    let target = availableResolutions[availableResolutions.length - 1];
    for (let i = 0; i < availableResolutions.length; i++) {
        if (availableResolutions[i] > spacetimeResolution) {
            target = availableResolutions[i];
            break;
        }
    }
    setResolution(target);
}

function setHeatmap2DResolution(target) {
    heatmap2DResolution = target;
    const cols = heatmap2DResolution;
    const rows = Math.max(8, Math.round(cols * (260 / 340)));
    const lbl = document.getElementById('lbl-heatmap-res');
    if (lbl) lbl.textContent = `${cols}x${rows}`;
    if (window.GravitySim && window.GravitySim.setConfig) {
        GravitySim.setConfig("heatmap_2d_resolution", heatmap2DResolution);
    }
    showToast(`2D Heatmap Resolution: ${cols}x${rows} (${cols * rows} Cells)`);
}

function decreaseHeatmap2DResolution() {
    let idx = availableHeatmapResolutions.indexOf(heatmap2DResolution);
    if (idx === -1) {
        idx = availableHeatmapResolutions.findIndex(r => r >= heatmap2DResolution);
        if (idx === -1) idx = availableHeatmapResolutions.length - 1;
    }
    if (idx > 0) {
        setHeatmap2DResolution(availableHeatmapResolutions[idx - 1]);
    } else {
        showToast(`2D Heatmap Resolution at minimum (${availableHeatmapResolutions[0]} cols)`);
    }
}

function increaseHeatmap2DResolution() {
    let idx = availableHeatmapResolutions.indexOf(heatmap2DResolution);
    if (idx === -1) {
        idx = availableHeatmapResolutions.findIndex(r => r >= heatmap2DResolution);
        if (idx === -1) idx = 0;
    }
    if (idx < availableHeatmapResolutions.length - 1) {
        setHeatmap2DResolution(availableHeatmapResolutions[idx + 1]);
    } else {
        showToast(`2D Heatmap Resolution at maximum (${availableHeatmapResolutions[availableHeatmapResolutions.length - 1]} cols)`);
    }
}

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

    // 3D Lagrange Equilibrium Points (L1-L5)
    initLagrangeVisualization();

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
    const steps = spacetimeResolution;
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
    spacetimeMesh.visible = showSpacetime || showGravitationalWaves;
    spacetimeMesh.userData.currentSpan = span;
    spacetimeMesh.userData.currentSteps = steps;
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
        if (window.GravitySim.getStats) {
            try {
                const stats = JSON.parse(GravitySim.getStats());
                const mergesElem = document.getElementById('hud-merges');
                if (mergesElem && stats.merges !== undefined) {
                    mergesElem.innerText = stats.merges;
                }
            } catch(e) {}
        }
    }

    // Update Gravitational Wave Strain HUD
    const hudGW = document.getElementById('hud-gw-item');
    if (hudGW) {
        if (showGravitationalWaves && window.GravitySim && window.GravitySim.getGWStats) {
            hudGW.classList.remove('hidden');
            try {
                const gwStats = JSON.parse(GravitySim.getGWStats());
                const strainElem = document.getElementById('hud-gw-val');
                if (strainElem) {
                    const strain = gwStats.strain || 0;
                    strainElem.innerText = strain > 1e-12 ? strain.toExponential(2) : "0.0e+00";
                    if (gwStats.bursts > 0) {
                        strainElem.innerText += ` (${gwStats.bursts} BURST)`;
                    }
                }
            } catch (e) {}
        } else {
            hudGW.classList.add('hidden');
        }
    }

    // Update 3D Visuals (Offloaded completely when 2D mode is active to maximize performance!)
    if (!offload3D) {
        updateCelestialBodies();
        updateSpacetimeGrid();
        updateVectorField();
        updateLagrangeVisualization();

        controls.update();
        renderer.render(scene, camera);
    }

    // Update 2D Tactical Viewport & Heatmap
    if (show2DViewport || offload3D) {
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

    // Update 3D body velocity direction vectors (respects showVectors flag!)
    if (showVectors && bodyCount > 0) {
        const maxVectors = bodyCount;
        if (!bodyVectorsLines || bodyVectorsLines.userData.maxVectors < maxVectors) {
            if (bodyVectorsLines) {
                scene.remove(bodyVectorsLines);
                bodyVectorsLines.geometry.dispose();
                bodyVectorsLines.material.dispose();
            }
            const posArr = new Float32Array(maxVectors * 6);
            const geo = new THREE.BufferGeometry();
            geo.setAttribute('position', new THREE.BufferAttribute(posArr, 3));
            const mat = new THREE.LineBasicMaterial({ color: 0x50ff80, transparent: true, opacity: 0.85 });
            bodyVectorsLines = new THREE.LineSegments(geo, mat);
            bodyVectorsLines.userData.maxVectors = maxVectors;
            scene.add(bodyVectorsLines);
        }
        bodyVectorsLines.visible = true;
        const posArr = bodyVectorsLines.geometry.attributes.position.array;
        let vIdx = 0;
        for (let i = 0; i < bodyCount; i++) {
            const off = i * stride;
            const flags = Math.round(buffer[off + 13]);
            const isStationary = (flags & 16) !== 0;
            if (isStationary) continue;
            const bx = buffer[off + 1];
            const by = buffer[off + 2];
            const bz = buffer[off + 3];
            const vx = buffer[off + 4];
            const vy = buffer[off + 5];
            const vz = buffer[off + 6];
            posArr[vIdx++] = bx;
            posArr[vIdx++] = by;
            posArr[vIdx++] = bz;
            posArr[vIdx++] = bx + vx * 1.5;
            posArr[vIdx++] = by + vy * 1.5;
            posArr[vIdx++] = bz + vz * 1.5;
        }
        bodyVectorsLines.geometry.setDrawRange(0, vIdx / 3);
        bodyVectorsLines.geometry.attributes.position.needsUpdate = true;
    } else if (bodyVectorsLines) {
        bodyVectorsLines.visible = false;
    }
}

// 5. Update 3D Spacetime Grid & Gravitational Waves
function updateSpacetimeGrid() {
    if ((!showSpacetime && !showGravitationalWaves) || !window.GravitySim) {
        if (spacetimeMesh) spacetimeMesh.visible = false;
        return;
    }

    const span = getStaticSpacetimeSpan();
    const steps = spacetimeResolution;
    if (!spacetimeMesh || Math.abs(spacetimeMesh.userData.currentSpan - span) > 0.01 || spacetimeMesh.userData.currentSteps !== steps) {
        createSpacetimeGridMesh();
    }
    spacetimeMesh.visible = true;

    const depths = GravitySim.getSpacetimeGrid(0, 0, span, steps);
    if (!depths || depths.length === 0) return;

    const posAttr = spacetimeMesh.geometry.attributes.position;
    const colorAttr = spacetimeMesh.geometry.attributes.color;
    const arr = posAttr.array;
    const cols = colorAttr.array;

    if (showSpacetime) {
        // Mode A: Spacetime Potential Wells (+ optional GW ripples modulating metric)
        spacetimeMesh.material.opacity = 0.40;
        for (let i = 0; i < depths.length; i++) {
            // In PlaneGeometry rotated -PI/2, Y is height
            const yVal = depths[i];
            arr[i * 3 + 1] = yVal;
            const depthRatio = Math.min(1.0, Math.abs(yVal) / 26.0);

            // Relativistic depth gradient: Deep cosmos blue -> Cyan/Teal -> Gold -> Fiery Orange/Red
            let r, g, b;
            if (depthRatio < 0.4) {
                const t = depthRatio / 0.4;
                r = 0.05 + t * 0.10;
                g = 0.35 + t * 0.45;
                b = 0.85 + t * 0.10;
            } else if (depthRatio < 0.75) {
                const t = (depthRatio - 0.4) / 0.35;
                r = 0.15 + t * 0.75;
                g = 0.80 + t * 0.10;
                b = 0.95 - t * 0.65;
            } else {
                const t = (depthRatio - 0.75) / 0.25;
                r = 0.90 + t * 0.10;
                g = 0.90 - t * 0.50;
                b = 0.30 - t * 0.20;
            }

            // Luminous fringe interference if GW ripples are also active
            if (showGravitationalWaves && yVal > -1.0) {
                r = Math.min(1.0, r + 0.15);
                b = Math.min(1.0, b + 0.25);
            }

            cols[i * 3 + 0] = r;
            cols[i * 3 + 1] = g;
            cols[i * 3 + 2] = b;
        }
    } else {
        // Mode B: Pure Gravitational Waves Mode (Spacetime Potential Wells OFF, GW Waves ON)
        // Flat baseline plane with undulating quadrupole spiral ripples & merger bursts
        spacetimeMesh.material.opacity = 0.60;
        for (let i = 0; i < depths.length; i++) {
            const yVal = depths[i];
            arr[i * 3 + 1] = yVal;
            const amp = Math.min(1.0, Math.abs(yVal) / 1.5);

            let r, g, b;
            if (yVal >= 0) {
                // Wave Crest: glowing electric cyan / neon turquoise
                r = 0.08 + amp * 0.15;
                g = 0.35 + amp * 0.60;
                b = 0.70 + amp * 0.30;
            } else {
                // Wave Trough: vibrant cosmic magenta / ultraviolet
                r = 0.35 + amp * 0.60;
                g = 0.10 + amp * 0.10;
                b = 0.65 + amp * 0.35;
            }

            cols[i * 3 + 0] = r;
            cols[i * 3 + 1] = g;
            cols[i * 3 + 2] = b;
        }
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

// 5c. 3D Lagrange Equilibrium Points (L1 - L5)
let lagrangeGroup = null;
let lagrangeMarkers = [];
let lagrangeLines = null;

function initLagrangeVisualization() {
    lagrangeGroup = new THREE.Group();
    lagrangeGroup.name = "lagrangeGroup";
    lagrangeGroup.visible = showLagrangePoints;

    const colors = [0xffdc32, 0x50dcff, 0xff7850, 0x78ffb4, 0xffb478];
    const names = ["L1", "L2", "L3", "L4", "L5"];

    for (let i = 0; i < 5; i++) {
        // Octahedron diamond marker
        const geo = new THREE.OctahedronGeometry(0.7, 0);
        const mat = new THREE.MeshBasicMaterial({
            color: colors[i],
            wireframe: true,
            transparent: true,
            opacity: 0.95
        });
        const marker = new THREE.Mesh(geo, mat);
        marker.userData = { index: i + 1, name: names[i] };
        lagrangeGroup.add(marker);
        lagrangeMarkers.push(marker);
    }

    // Dashed guide lines
    const lineMat = new THREE.LineBasicMaterial({
        color: 0x4080bb,
        transparent: true,
        opacity: 0.45
    });
    const lineGeo = new THREE.BufferGeometry();
    lineGeo.setAttribute('position', new THREE.BufferAttribute(new Float32Array(12 * 3), 3));
    lagrangeLines = new THREE.LineSegments(lineGeo, lineMat);
    lagrangeGroup.add(lagrangeLines);

    scene.add(lagrangeGroup);
}

function updateLagrangeVisualization() {
    if (!showLagrangePoints || !window.GravitySim || !lagrangeGroup) {
        if (lagrangeGroup) lagrangeGroup.visible = false;
        return;
    }

    const lDataStr = GravitySim.getLagrangePoints();
    if (!lDataStr || lDataStr === "{}") {
        lagrangeGroup.visible = false;
        return;
    }

    try {
        const lData = JSON.parse(lDataStr);
        if (!lData.points || lData.points.length < 5) {
            lagrangeGroup.visible = false;
            return;
        }

        lagrangeGroup.visible = true;
        const pts = lData.points;
        const now = performance.now() * 0.002;

        for (let i = 0; i < 5; i++) {
            const p = pts[i];
            const marker = lagrangeMarkers[i];
            marker.position.set(p.pos.x, p.pos.y, p.pos.z);
            marker.rotation.x = now + i;
            marker.rotation.y = now * 1.4 + i;
            const s = 1.0 + 0.18 * Math.sin(now * 3.5 + i);
            marker.scale.set(s, s, s);
        }

        if (lagrangeLines && lData.primary && lData.secondary) {
            const pos = lagrangeLines.geometry.attributes.position.array;
            const p1 = lData.primary.pos;
            const p2 = lData.secondary.pos;
            const l4 = pts[3].pos;
            const l5 = pts[4].pos;

            // Primary -> Secondary
            pos[0] = p1.x; pos[1] = p1.y; pos[2] = p1.z;
            pos[3] = p2.x; pos[4] = p2.y; pos[5] = p2.z;
            // Primary -> L4
            pos[6] = p1.x; pos[7] = p1.y; pos[8] = p1.z;
            pos[9] = l4.x; pos[10] = l4.y; pos[11] = l4.z;
            // Secondary -> L4
            pos[12] = p2.x; pos[13] = p2.y; pos[14] = p2.z;
            pos[15] = l4.x; pos[16] = l4.y; pos[17] = l4.z;
            // Primary -> L5
            pos[18] = p1.x; pos[19] = p1.y; pos[20] = p1.z;
            pos[21] = l5.x; pos[22] = l5.y; pos[23] = l5.z;
            // Secondary -> L5
            pos[24] = p2.x; pos[25] = p2.y; pos[26] = p2.z;
            pos[27] = l5.x; pos[28] = l5.y; pos[29] = l5.z;
            // L1 -> L2
            pos[30] = pts[0].pos.x; pos[31] = pts[0].pos.y; pos[32] = pts[0].pos.z;
            pos[33] = pts[1].pos.x; pos[34] = pts[1].pos.y; pos[35] = pts[1].pos.z;

            lagrangeLines.geometry.attributes.position.needsUpdate = true;
        }
    } catch (e) {
        lagrangeGroup.visible = false;
    }
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
        setTrailsEnabled(!showTrails);
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
        if (window.GravitySim && window.GravitySim.setConfig) {
            GravitySim.setConfig("show_potential_grid", showSpacetime);
        }
        if (spacetimeMesh) {
            spacetimeMesh.visible = showSpacetime || showGravitationalWaves;
        }
        showToast(`Spacetime Curvature Grid: ${showSpacetime ? 'ON' : 'OFF'}`);
    });

    const togGWWaves = document.getElementById('tog-gwwaves');
    if (togGWWaves) {
        togGWWaves.addEventListener('click', () => {
            showGravitationalWaves = !showGravitationalWaves;
            togGWWaves.classList.toggle('active', showGravitationalWaves);
            if (window.GravitySim && window.GravitySim.setConfig) {
                GravitySim.setConfig("show_gravitational_waves", showGravitationalWaves);
            }
            if (spacetimeMesh) {
                spacetimeMesh.visible = showSpacetime || showGravitationalWaves;
            }
            showToast(`Gravitational Waves & LIGO: ${showGravitationalWaves ? 'ON' : 'OFF'}`);
        });
    }

    const togVectors = document.getElementById('tog-vectors');
    if (togVectors) {
        togVectors.addEventListener('click', () => {
            setVectorsEnabled(!showVectors);
        });
    }

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
            if (lagrangeGroup) lagrangeGroup.visible = showLagrangePoints;
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
        } else if (e.key === '{') {
            if (e.altKey || (show2DHeatmap && offload3D)) {
                decreaseHeatmap2DResolution();
            } else {
                decreaseResolution();
            }
        } else if (e.key === '}') {
            if (e.altKey || (show2DHeatmap && offload3D)) {
                increaseHeatmap2DResolution();
            } else {
                increaseResolution();
            }
        } else if (e.key === 'c' || e.key === 'C') {
            set2DCirclesEnabled(!show2DCircles);
        } else if (e.key === '[') {
            if (e.altKey) {
                decreaseHeatmap2DResolution();
            } else if (e.shiftKey) {
                decreaseResolution();
            } else {
                spacetimeScale = Math.max(0.5, spacetimeScale - 0.25);
                createSpacetimeGridMesh();
                showToast(`Spacetime Scale: ${spacetimeScale.toFixed(2)}x`);
            }
        } else if (e.key === ']') {
            if (e.altKey) {
                increaseHeatmap2DResolution();
            } else if (e.shiftKey) {
                increaseResolution();
            } else {
                spacetimeScale = Math.min(3.5, spacetimeScale + 0.25);
                createSpacetimeGridMesh();
                showToast(`Spacetime Scale: ${spacetimeScale.toFixed(2)}x`);
            }
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
        } else if (e.key === 'z' || e.key === 'Z' || e.key === 'F5') {
            e.preventDefault();
            setOffload3D(!offload3D);
        } else if (e.key === 'f' || e.key === 'F') {
            focus2D();
        } else if (e.key === 't' || e.key === 'T') {
            setTrack2D(!track2D);
        } else if (e.key === 'k' || e.key === 'K') {
            const btn = document.getElementById('tog-2dlabels');
            if (btn) btn.click();
        } else if (e.key === 'y' || e.key === 'Y') {
            const btn = document.getElementById('btn-2d-icons') || document.getElementById('tog-2dicons');
            if (btn) btn.click();
        } else if (e.key === 'u' || e.key === 'U') {
            const btn = document.getElementById('tog-lagrange');
            if (btn) btn.click();
        } else if (e.key === 'j' || e.key === 'J') {
            const btn = document.getElementById('tog-gwwaves');
            if (btn) btn.click();
        } else if (e.key === 'x' || e.key === 'X') {
            const btn = document.getElementById('btn-2d-probe');
            if (btn) btn.click();
        } else if (e.key === 'v' || e.key === 'V') {
            setVectorsEnabled(!showVectors);
        } else if (e.key === 'i' || e.key === 'I' || e.key === 'F11' || e.key === 'F3') {
            const hud = document.getElementById('hud-overlay');
            const btn2dHUD = document.getElementById('btn-2d-hud');
            if (hud) {
                hud.classList.toggle('hidden');
                const isHidden = hud.classList.contains('hidden');
                if (btn2dHUD) btn2dHUD.classList.toggle('active', !isHidden);
                showToast(isHidden ? 'HUD Telemetry: OFF' : 'HUD Telemetry: ON');
            }
        } else if (e.key === 'h' || e.key === 'H') {
            if (e.shiftKey) {
                increaseHeatmap2DResolution();
            } else if (e.altKey || e.ctrlKey) {
                decreaseHeatmap2DResolution();
            } else {
                const btnH = document.getElementById('tog-heatmap');
                if (btnH) btnH.click();
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

    if (showSpacetime || showGravitationalWaves) {
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
            executeSpawn(spawnStartPos, drag);
        }
        spawnStartPos = null;
        spawnAimLine.visible = false;
        controls.enabled = true;
    }
}

function executeSpawn(pos, drag) {
    if (!window.GravitySim) return;
    const spawnType = document.getElementById('select-spawn').value;

    if (spawnType === 'swarm50') {
        GravitySim.spawnCluster("cluster50", pos.x, pos.y, pos.z, drag.x, drag.y, drag.z);
    } else if (spawnType === 'collapse100') {
        GravitySim.spawnCluster("collapse100", pos.x, pos.y, pos.z, drag.x, drag.y, drag.z);
    } else if (spawnType === 'galaxy150') {
        GravitySim.spawnCluster("galaxy150", pos.x, pos.y, pos.z, drag.x, drag.y, drag.z);
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
            pos.x, pos.y, pos.z,
            drag.x, drag.y, drag.z,
            mass, rad, r, g, b, isStar, isBH);
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
let show2DIcons = false;
let show2DHeatmap = true;
let show3DHeatmapPlane = false;
let showLagrangePoints = true;
let viewport2DFullscreen = false;
let viewport2DZoom = 0.5;
let viewport2DPan = { x: 0, z: 0 };
let isDragging2D = false;
let lastMouse2D = { x: 0, y: 0 };

let offload3D = false;
let track2D = false;
let isSpawning2D = false;
let spawnStart2D = null;
let spawnCur2D = null;
let show2DGW = true;

function setOffload3D(enable) {
    offload3D = enable;
    const container = document.getElementById('viewport-2d-container');
    const togTop = document.getElementById('btn-2d-offload-top');
    const togBottom = document.getElementById('tog-2doffload');
    const btnHeader = document.getElementById('btn-2d-offload');
    const titleEl = document.getElementById('viewport-2d-title');

    if (offload3D) {
        show2DViewport = true;
        if (container) {
            container.classList.remove('hidden');
            container.classList.add('offload-mode');
        }
        if (titleEl) titleEl.innerText = "🛰️ 2D MAIN COMMAND VIEWPORT · [3D OFFLOADED]";
        showToast("⚡ 3D Render Offloaded: Running purely on 2D Main Viewport");
    } else {
        if (container) {
            container.classList.remove('offload-mode');
        }
        if (titleEl) titleEl.innerText = "🛰️ 2D TACTICAL ORBITAL MAP";
        showToast("🎥 3D Render Restored: Standard Mode");
    }
    document.body.classList.toggle('offload-active', offload3D);

    if (togTop) {
        togTop.classList.toggle('active', offload3D);
        togTop.innerText = offload3D ? "⚡ 3D OFF (2D)" : "⚡ 2D OFFLOAD";
    }
    if (togBottom) {
        togBottom.classList.toggle('active', offload3D);
    }
    if (btnHeader) {
        btnHeader.classList.toggle('active', offload3D);
        btnHeader.innerText = offload3D ? "3D OFF" : "3D Offload";
    }
    update2DCanvasSize();
}

function focus2D() {
    if (!window.GravitySim) return;
    if (selectedBodyId > 0) {
        const buffer = GravitySim.getBodiesBuffer();
        const stride = 16;
        const count = buffer.length / stride;
        for (let i = 0; i < count; i++) {
            const off = i * stride;
            if (Math.round(buffer[off + 0]) === selectedBodyId) {
                viewport2DPan.x = buffer[off + 1];
                viewport2DPan.z = buffer[off + 3];
                showToast("2D Centered on Selected Object");
                return;
            }
        }
    }
    const buffer = GravitySim.getBodiesBuffer();
    const stride = 16;
    const count = buffer.length / stride;
    let maxMass = -1, bestX = 0, bestZ = 0;
    for (let i = 0; i < count; i++) {
        const off = i * stride;
        const mass = buffer[off + 8];
        if (mass > maxMass) {
            maxMass = mass;
            bestX = buffer[off + 1];
            bestZ = buffer[off + 3];
        }
    }
    if (maxMass > 0) {
        viewport2DPan.x = bestX;
        viewport2DPan.z = bestZ;
        showToast("2D Centered on Primary Heavy Mass");
    }
}

function setTrack2D(enable) {
    track2D = enable;
    const btnTrack = document.getElementById('btn-2d-track');
    if (btnTrack) {
        btnTrack.classList.toggle('active', track2D);
        btnTrack.innerText = track2D ? "TRACK ON" : "Track";
    }
    showToast(`2D Target Tracking: ${track2D ? 'LOCKED ON' : 'OFF'}`);
    if (track2D) focus2D();
}

function setSpawn2D(enable) {
    isSpawning2D = enable;
    spawnStart2D = null;
    spawnCur2D = null;
    const btnSpawn = document.getElementById('btn-2d-spawn');
    if (btnSpawn) {
        btnSpawn.classList.toggle('active', isSpawning2D);
        btnSpawn.innerText = isSpawning2D ? "SPAWN ON" : "+Spawn";
    }
    showToast(`2D Spawn Mode: ${isSpawning2D ? 'ON (Click & Drag slingshot to launch)' : 'OFF'}`);
}

function setup2DViewport() {
    const container = document.getElementById('viewport-2d-container');
    const canvas = document.getElementById('canvas-2d');
    const tog2DMap = document.getElementById('tog-2dmap');
    const togOffload = document.getElementById('tog-2doffload');
    const btnOffloadTop = document.getElementById('btn-2d-offload-top');
    const btnOffloadHeader = document.getElementById('btn-2d-offload');
    const btnFocus = document.getElementById('btn-2d-focus');
    const btnTrack = document.getElementById('btn-2d-track');
    const btnSpawn = document.getElementById('btn-2d-spawn');
    const btnGW = document.getElementById('btn-2d-gw');
    const togHeatmap = document.getElementById('tog-heatmap');
    const btnLabels = document.getElementById('btn-2d-labels');
    const btnIcons = document.getElementById('btn-2d-icons');
    const togIcons = document.getElementById('tog-2dicons');
    const btnProbe = document.getElementById('btn-2d-probe');
    const btnReset = document.getElementById('btn-2d-reset');
    const btnFullscreen = document.getElementById('btn-2d-fullscreen');
    const btnClose = document.getElementById('btn-close-2d');

    if (!container || !canvas) return;

    if (btnOffloadTop) {
        btnOffloadTop.addEventListener('click', () => setOffload3D(!offload3D));
    }
    if (togOffload) {
        togOffload.addEventListener('click', () => setOffload3D(!offload3D));
    }
    if (btnOffloadHeader) {
        btnOffloadHeader.addEventListener('click', () => setOffload3D(!offload3D));
    }

    if (btnFocus) {
        btnFocus.addEventListener('click', focus2D);
    }
    if (btnTrack) {
        btnTrack.addEventListener('click', () => setTrack2D(!track2D));
    }
    if (btnSpawn) {
        btnSpawn.addEventListener('click', () => setSpawn2D(!isSpawning2D));
    }
    if (btnGW) {
        btnGW.addEventListener('click', () => {
            show2DGW = !show2DGW;
            btnGW.classList.toggle('active', show2DGW);
            showToast(`2D Gravitational Waves: ${show2DGW ? 'ON' : 'OFF'}`);
        });
    }

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
            const togL = document.getElementById('tog-2dlabels');
            if (togL) togL.classList.toggle('active', show2DLabels);
            showToast(`2D Labels: ${show2DLabels ? 'ON' : 'OFF (Clean Mode)'}`);
        });
    }

    if (btnIcons) {
        btnIcons.addEventListener('click', () => {
            show2DIcons = !show2DIcons;
            btnIcons.classList.toggle('active', show2DIcons);
            if (togIcons) togIcons.classList.toggle('active', show2DIcons);
            showToast(`2D Celestial Icons: ${show2DIcons ? 'ON (Glyphs)' : 'OFF (Clear Circles)'}`);
        });
    }

    if (togIcons) {
        togIcons.addEventListener('click', () => {
            show2DIcons = !show2DIcons;
            togIcons.classList.toggle('active', show2DIcons);
            if (btnIcons) btnIcons.classList.toggle('active', show2DIcons);
            showToast(`2D Celestial Icons: ${show2DIcons ? 'ON (Glyphs)' : 'OFF (Clear Circles)'}`);
        });
    }

    if (btnProbe) {
        btnProbe.addEventListener('click', () => {
            if (window.GravitySim) {
                const probeId = GravitySim.spawnLagrangeProbe(4);
                if (probeId > 0) {
                    showToast("🚀 Scientific Satellite Probe deployed to L4 Trojan Point!");
                } else {
                    showToast("No dominant 2-body pair found for Lagrange points.");
                }
            }
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
            document.body.classList.toggle('viewport-2d-fullscreen', viewport2DFullscreen);
            update2DCanvasSize();
        });
    }

    if (btnClose) {
        btnClose.addEventListener('click', () => {
            show2DViewport = false;
            if (offload3D) setOffload3D(false);
            if (tog2DMap) tog2DMap.classList.remove('active');
            container.classList.add('hidden');
        });
    }

    const btn2dVec = document.getElementById('btn-2d-vec');
    if (btn2dVec) {
        btn2dVec.addEventListener('click', () => {
            setVectorsEnabled(!showVectors);
        });
    }

    const btn2dHUD = document.getElementById('btn-2d-hud');
    if (btn2dHUD) {
        btn2dHUD.addEventListener('click', () => {
            const hud = document.getElementById('hud-overlay');
            if (hud) {
                hud.classList.toggle('hidden');
                const isHidden = hud.classList.contains('hidden');
                btn2dHUD.classList.toggle('active', !isHidden);
                showToast(isHidden ? 'HUD Telemetry: OFF' : 'HUD Telemetry: ON');
            }
        });
    }

    const btnCloseHUD = document.getElementById('btn-close-hud');
    if (btnCloseHUD) {
        btnCloseHUD.addEventListener('click', () => {
            const hud = document.getElementById('hud-overlay');
            if (hud) {
                hud.classList.add('hidden');
                if (btn2dHUD) btn2dHUD.classList.remove('active');
                showToast('HUD Telemetry: OFF (Press I/F11 to restore)');
            }
        });
    }

    const btn2dTrl = document.getElementById('btn-2d-trl');
    if (btn2dTrl) {
        btn2dTrl.addEventListener('click', () => {
            setTrailsEnabled(!showTrails);
        });
    }

    const btn2dCir = document.getElementById('btn-2d-cir');
    if (btn2dCir) {
        btn2dCir.addEventListener('click', () => {
            set2DCirclesEnabled(!show2DCircles);
        });
    }

    const tog2dCircles = document.getElementById('tog-2dcircles');
    if (tog2dCircles) {
        tog2dCircles.addEventListener('click', () => {
            set2DCirclesEnabled(!show2DCircles);
        });
    }

    const btn2dResDec = document.getElementById('btn-2d-res-dec');
    if (btn2dResDec) {
        btn2dResDec.addEventListener('click', (e) => {
            if (show2DHeatmap && !e.shiftKey) {
                decreaseHeatmap2DResolution();
            } else {
                decreaseResolution();
            }
        });
    }
    const btn2dResInc = document.getElementById('btn-2d-res-inc');
    if (btn2dResInc) {
        btn2dResInc.addEventListener('click', (e) => {
            if (show2DHeatmap && !e.shiftKey) {
                increaseHeatmap2DResolution();
            } else {
                increaseResolution();
            }
        });
    }
    const btnResDec = document.getElementById('btn-res-dec');
    if (btnResDec) {
        btnResDec.addEventListener('click', decreaseResolution);
    }
    const btnResInc = document.getElementById('btn-res-inc');
    if (btnResInc) {
        btnResInc.addEventListener('click', increaseResolution);
    }
    const btnHResDec = document.getElementById('btn-hres-dec');
    if (btnHResDec) {
        btnHResDec.addEventListener('click', decreaseHeatmap2DResolution);
    }
    const btnHResInc = document.getElementById('btn-hres-inc');
    if (btnHResInc) {
        btnHResInc.addEventListener('click', increaseHeatmap2DResolution);
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
        if (isSpawning2D) {
            const rect = canvas.getBoundingClientRect();
            const sx = e.clientX - rect.left;
            const sy = e.clientY - rect.top;
            const cx = canvas.width * 0.5;
            const cy = canvas.height * 0.5;
            spawnStart2D = {
                x: viewport2DPan.x + (sx - cx) / viewport2DZoom,
                z: viewport2DPan.z + (sy - cy) / viewport2DZoom
            };
            spawnCur2D = { ...spawnStart2D };
            return;
        }

        if (e.button === 0) {
            // Select object on 2D map
            const rect = canvas.getBoundingClientRect();
            const sx = e.clientX - rect.left;
            const sy = e.clientY - rect.top;
            const cx = canvas.width * 0.5;
            const cy = canvas.height * 0.5;
            const clickWorldX = viewport2DPan.x + (sx - cx) / viewport2DZoom;
            const clickWorldZ = viewport2DPan.z + (sy - cy) / viewport2DZoom;

            if (window.GravitySim) {
                const buffer = GravitySim.getBodiesBuffer();
                const stride = 16;
                const count = buffer.length / stride;
                let closestId = -1;
                let closestDist = 25 / viewport2DZoom;
                for (let i = 0; i < count; i++) {
                    const off = i * stride;
                    const id = Math.round(buffer[off + 0]);
                    const bx = buffer[off + 1];
                    const bz = buffer[off + 3];
                    const d = Math.hypot(bx - clickWorldX, bz - clickWorldZ);
                    if (d < closestDist) {
                        closestDist = d;
                        closestId = id;
                    }
                }
                if (closestId > 0) {
                    selectedBodyId = closestId;
                    GravitySim.selectBody(closestId);
                    document.getElementById('inspector-panel').classList.remove('hidden');
                    document.body.classList.add('inspector-active');
                    showToast(`Selected Body #${closestId} on 2D Map`);
                    return;
                }
            }
        }

        if (e.button === 2 || e.button === 0) {
            isDragging2D = true;
            lastMouse2D = { x: e.clientX, y: e.clientY };
            track2D = false;
        }
    });

    window.addEventListener('mousemove', (e) => {
        if (isSpawning2D && spawnStart2D) {
            const rect = canvas.getBoundingClientRect();
            const sx = e.clientX - rect.left;
            const sy = e.clientY - rect.top;
            const cx = canvas.width * 0.5;
            const cy = canvas.height * 0.5;
            spawnCur2D = {
                x: viewport2DPan.x + (sx - cx) / viewport2DZoom,
                z: viewport2DPan.z + (sy - cy) / viewport2DZoom
            };
            return;
        }
        if (isDragging2D) {
            const dx = e.clientX - lastMouse2D.x;
            const dy = e.clientY - lastMouse2D.y;
            viewport2DPan.x -= dx / viewport2DZoom;
            viewport2DPan.z -= dy / viewport2DZoom;
            lastMouse2D = { x: e.clientX, y: e.clientY };
        }
    });

    window.addEventListener('mouseup', () => {
        if (isSpawning2D && spawnStart2D && spawnCur2D) {
            const drag = {
                x: (spawnStart2D.x - spawnCur2D.x) * 0.25,
                y: 0,
                z: (spawnStart2D.z - spawnCur2D.z) * 0.25
            };
            executeSpawn({ x: spawnStart2D.x, y: 0, z: spawnStart2D.z }, drag);
            showToast("🚀 Spawned object into 2D orbit!");
            spawnStart2D = null;
            spawnCur2D = null;
            return;
        }
        isDragging2D = false;
    });

    canvas.addEventListener('contextmenu', (e) => e.preventDefault());
    window.addEventListener('resize', update2DCanvasSize);
}

function update2DCanvasSize() {
    const container = document.getElementById('viewport-2d-container');
    const canvas = document.getElementById('canvas-2d');
    if (!container || !canvas) return;
    if (offload3D) {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight - 50 - 48 - 32;
    } else if (viewport2DFullscreen) {
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

    // Continuous 2D Target Tracking Update
    if (track2D && selectedBodyId > 0) {
        for (let i = 0; i < bodyCount; i++) {
            const off = i * stride;
            if (Math.round(buffer[off + 0]) === selectedBodyId) {
                viewport2DPan.x = buffer[off + 1];
                viewport2DPan.z = buffer[off + 3];
                break;
            }
        }
    }

    // 2. Gravitational Heatmap
    if (show2DHeatmap && bodyCount > 0) {
        const cellCols = heatmap2DResolution || 64;
        const cellRows = Math.max(8, Math.round(cellCols * (h / w)));
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

    // 3b. Render Lagrange Points L1..L5 in 2D
    if (showLagrangePoints && window.GravitySim) {
        const lDataStr = GravitySim.getLagrangePoints();
        if (lDataStr && lDataStr !== "{}") {
            try {
                const lData = JSON.parse(lDataStr);
                if (lData.points && lData.points.length >= 5) {
                    const lColors = [
                        '#ffdc32', // L1 Gold
                        '#50dcff', // L2 Cyan
                        '#ff7850', // L3 Coral
                        '#78ffb4', // L4 Emerald
                        '#ffb478'  // L5 Amber
                    ];
                    if (lData.primary && lData.secondary) {
                        const p1 = lData.primary.pos;
                        const p2 = lData.secondary.pos;
                        const sx1 = toScreenX(p1.x), sy1 = toScreenY(p1.z);
                        const sx2 = toScreenX(p2.x), sy2 = toScreenY(p2.z);

                        ctx.strokeStyle = 'rgba(60, 120, 190, 0.4)';
                        ctx.setLineDash([3, 3]);
                        ctx.beginPath();
                        ctx.moveTo(sx1, sy1); ctx.lineTo(sx2, sy2);
                        ctx.stroke();
                        ctx.setLineDash([]);
                    }

                    for (let i = 0; i < 5; i++) {
                        const pt = lData.points[i];
                        const lsx = toScreenX(pt.pos.x);
                        const lsy = toScreenY(pt.pos.z);
                        if (lsx < -20 || lsx > w + 20 || lsy < -20 || lsy > h + 20) continue;

                        const col = lColors[i];
                        // Diamond marker
                        ctx.save();
                        ctx.translate(lsx, lsy);
                        ctx.rotate(Math.PI / 4);
                        ctx.strokeStyle = col;
                        ctx.lineWidth = 1.5;
                        ctx.strokeRect(-3.5, -3.5, 7, 7);
                        ctx.fillStyle = col;
                        ctx.fillRect(-1, -1, 2, 2);
                        ctx.restore();

                        // Label
                        ctx.fillStyle = col;
                        ctx.font = 'bold 9px sans-serif';
                        ctx.fillText(`L${i + 1}`, lsx + 8, lsy + 3);
                    }
                }
            } catch (err) {}
        }
    }

    // 3c. Gravitational Waves ripples & shockwaves in 2D
    if (showGravitationalWaves && window.GravitySim && window.GravitySim.getGravitationalWaves) {
        try {
            const gwStatsStr = GravitySim.getGravitationalWaves();
            const gwStats = JSON.parse(gwStatsStr || "{}");
            const curTime = performance.now() * 0.001;

            // Continuous binary ripples around heaviest pair
            if (bodyCount >= 2) {
                let p1x = 0, p1z = 0, m1 = -1;
                let p2x = 0, p2z = 0, m2 = -1;
                for (let i = 0; i < Math.min(bodyCount, 30); i++) {
                    const off = i * stride;
                    const mass = buffer[off + 8];
                    if (mass > m1) {
                        m2 = m1; p2x = p1x; p2z = p1z;
                        m1 = mass; p1x = buffer[off + 1]; p1z = buffer[off + 3];
                    } else if (mass > m2) {
                        m2 = mass; p2x = buffer[off + 1]; p2z = buffer[off + 3];
                    }
                }
                if (m1 > 0 && m2 > 0.0005) {
                    const baryX = (p1x * m1 + p2x * m2) / (m1 + m2);
                    const baryZ = (p1z * m1 + p2z * m2) / (m1 + m2);
                    const bsx = toScreenX(baryX);
                    const bsy = toScreenY(baryZ);

                    const c = 150.0;
                    const fEff = Math.max(0.05, gwStats.freq || 0.5);
                    const gwWavelength = Math.max(12.0, (c / fEff) * 0.5);
                    const phaseOffset = (curTime * c * 0.5) % gwWavelength;

                    ctx.lineWidth = 1.3;
                    for (let ring = 0; ring < 7; ring++) {
                        const rWorld = ring * gwWavelength + phaseOffset;
                        const rPix = rWorld * zoom;
                        if (rPix > 5 && rPix < Math.max(w, h) * 1.5) {
                            const alpha = Math.max(0, 0.5 * (1.0 - ring / 7));
                            ctx.strokeStyle = `rgba(190, 80, 255, ${alpha.toFixed(2)})`;
                            ctx.beginPath();
                            ctx.arc(bsx, bsy, rPix, 0, Math.PI * 2);
                            ctx.stroke();
                        }
                    }
                }
            }

            // Expanding merger bursts shockwaves
            if (gwStats.active_bursts && gwStats.active_bursts.length > 0) {
                for (const b of gwStats.active_bursts) {
                    const age = curTime - b.start_time;
                    if (age > 0 && age <= b.decay_time) {
                        const bsx = toScreenX(b.pos.x);
                        const bsy = toScreenY(b.pos.z);
                        const rPix = b.wave_speed * age * zoom;
                        const intensity = Math.exp(-age * 0.75);
                        const alpha = Math.max(0, intensity * 0.85);
                        if (alpha > 0.05) {
                            ctx.strokeStyle = `rgba(255, 100, 200, ${alpha.toFixed(2)})`;
                            ctx.lineWidth = 2.2;
                            ctx.beginPath();
                            ctx.arc(bsx, bsy, rPix, 0, Math.PI * 2);
                            ctx.stroke();
                        }
                    }
                }
            }

            // 2D GW Telemetry Pill Badge
            const strainVal = gwStats.strain || 0;
            const freqVal = gwStats.freq || 0;
            if (strainVal > 1e-12 || freqVal > 0) {
                const gwBadgeText = `GW · h: ${strainVal > 1e-12 ? strainVal.toExponential(2) : "0.0e+00"} · f: ${freqVal.toFixed(1)} Hz`;
                ctx.font = 'bold 9px sans-serif';
                ctx.textAlign = 'left';
                const tw = ctx.measureText(gwBadgeText).width;
                ctx.fillStyle = 'rgba(24, 10, 36, 0.85)';
                ctx.fillRect(10, 10, tw + 14, 18);
                ctx.strokeStyle = 'rgba(180, 80, 240, 0.7)';
                ctx.lineWidth = 1;
                ctx.strokeRect(10, 10, tw + 14, 18);
                ctx.fillStyle = '#e6a0ff';
                ctx.fillText(gwBadgeText, 17, 22);
            }
        } catch (err) {}
    }

    // 3d. Orbit Reference Trails in 2D (Respects showTrails toggle)
    if (showTrails && trailsMap.size > 0) {
        for (const [id, trail] of trailsMap.entries()) {
            if (!trail || !trail.geometry) continue;
            const posAttr = trail.geometry.attributes.position;
            const maxPts = trail.userData ? trail.userData.maxPts : 0;
            const count = trail.userData ? trail.userData.count : 0;
            const head = trail.userData ? trail.userData.head : 0;
            if (count > 1 && maxPts > 0) {
                const trailCol = trail.material && trail.material.color ? trail.material.color.getStyle() : 'rgba(100, 200, 255, 0.6)';
                ctx.strokeStyle = trailCol;
                ctx.lineWidth = 1.2;
                ctx.beginPath();
                let started = false;
                for (let i = 0; i < count; i++) {
                    const idx = (head - count + i + maxPts) % maxPts;
                    const tx = posAttr.getX(idx);
                    const tz = posAttr.getZ(idx);
                    const sx = toScreenX(tx);
                    const sy = toScreenY(tz);
                    if (!started) {
                        ctx.moveTo(sx, sy);
                        started = true;
                    } else {
                        ctx.lineTo(sx, sy);
                    }
                }
                ctx.stroke();
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
        const isBlackHole = (flags & 2) !== 0;

        const celestialIcon = Math.round(buffer[off + 14]);

        const sx = toScreenX(bx);
        const sy = toScreenY(bz);

        if (sx < -40 || sx > w + 40 || sy < -40 || sy > h + 40) continue;

        let drawR = Math.max(2.0, Math.min(18.0, Math.log10(mass + 1.0) * 3.0 * zoom + 1.5));

        // Velocity vector indicator (respects showVectors flag, like 3D!)
        if (showVectors) {
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
        }

        // Body visual (Clear circle mode default, icons mode when toggled)
        if (show2DIcons) {
            let glyph = '';
            switch (celestialIcon) {
                case 1: glyph = '☀️'; break; // Star
                case 2: glyph = '🕳️'; break; // Black Hole
                case 3: glyph = '⚡'; break; // Pulsar
                case 4: glyph = '🪐'; break; // Gas Giant
                case 5: glyph = '🌍'; break; // Terrestrial
                case 6: glyph = '🔴'; break; // Desert
                case 7: glyph = '❄️'; break; // Ice Giant
                case 8: glyph = '🌙'; break; // Moon
                case 9: glyph = '🛰️'; break; // Probe
            }

            if (glyph) {
                // Show just icon (no body disc / object picture)
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                const fontSize = Math.max(12, Math.round(drawR * 1.5));
                ctx.font = `${fontSize}px sans-serif`;
                ctx.fillText(glyph, sx, sy);
            } else {
                ctx.strokeStyle = `rgb(${cr},${cg},${cb})`;
                ctx.lineWidth = 1.5;
                ctx.strokeRect(sx - 3, sy - 3, 6, 6);
            }
        } else {
            // Clear circle & dot mode: pristine outline circle with translucent center core dot (both toggle together)
            if (show2DCircles) {
                ctx.strokeStyle = `rgb(${cr},${cg},${cb})`;
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.arc(sx, sy, drawR, 0, Math.PI * 2);
                ctx.stroke();

                // Center core dot for pinpoint accuracy
                ctx.fillStyle = `rgba(${cr},${cg},${cb}, 0.85)`;
                ctx.beginPath();
                ctx.arc(sx, sy, 1.2, 0, Math.PI * 2);
                ctx.fill();
            }
        }

        // Label for heavy or selected bodies (controlled by show2DLabels toggle)
        if (show2DLabels && (mass >= 5.0 || (bodyCount < 30 && mass > 0.1))) {
            ctx.fillStyle = 'rgba(220, 235, 255, 0.85)';
            ctx.fillText(`M:${mass >= 100 ? mass.toFixed(0) : mass.toFixed(1)}`, sx, sy - drawR - 3);
        }
    }

    // 5. Target Lock Reticle
    if (track2D && selectedBodyId > 0) {
        for (let i = 0; i < bodyCount; i++) {
            const off = i * stride;
            if (Math.round(buffer[off + 0]) === selectedBodyId) {
                const sx = toScreenX(buffer[off + 1]);
                const sy = toScreenY(buffer[off + 3]);
                const mass = buffer[off + 8];
                const drawR = Math.max(2.0, Math.min(18.0, Math.log10(mass + 1.0) * 3.0 * zoom + 1.5)) + 6;
                const pulse = Math.sin(performance.now() * 0.006) * 2.0 + 2.0;
                const rLock = drawR + pulse;

                ctx.strokeStyle = '#14f0c8';
                ctx.lineWidth = 1.5;
                const bLen = 6;

                // Corner brackets
                ctx.beginPath();
                ctx.moveTo(sx - rLock, sy - rLock + bLen);
                ctx.lineTo(sx - rLock, sy - rLock);
                ctx.lineTo(sx - rLock + bLen, sy - rLock);
                ctx.stroke();

                ctx.beginPath();
                ctx.moveTo(sx + rLock - bLen, sy - rLock);
                ctx.lineTo(sx + rLock, sy - rLock);
                ctx.lineTo(sx + rLock, sy - rLock + bLen);
                ctx.stroke();

                ctx.beginPath();
                ctx.moveTo(sx - rLock, sy + rLock - bLen);
                ctx.lineTo(sx - rLock, sy + rLock);
                ctx.lineTo(sx - rLock + bLen, sy + rLock);
                ctx.stroke();

                ctx.beginPath();
                ctx.moveTo(sx + rLock - bLen, sy + rLock);
                ctx.lineTo(sx + rLock, sy + rLock);
                ctx.lineTo(sx + rLock, sy + rLock - bLen);
                ctx.stroke();

                ctx.fillStyle = '#14f0c8';
                ctx.font = 'bold 9px sans-serif';
                ctx.textAlign = 'center';
                ctx.fillText('⛶ TRACK LOCKED', sx, sy + rLock + 11);
                break;
            }
        }
    }

    // 6. 2D Spawning Trajectory Overlay
    if (isSpawning2D) {
        const spawnType = document.getElementById('select-spawn').value;
        const bannerText = `▶ 2D SPAWN ACTIVE: Click & Drag slingshot to launch (${spawnType})`;
        ctx.font = 'bold 10px sans-serif';
        ctx.textAlign = 'center';
        const bw = ctx.measureText(bannerText).width;
        ctx.fillStyle = 'rgba(20, 28, 48, 0.88)';
        ctx.fillRect(cx - bw * 0.5 - 10, 10, bw + 20, 20);
        ctx.strokeStyle = 'rgba(255, 140, 40, 0.9)';
        ctx.lineWidth = 1;
        ctx.strokeRect(cx - bw * 0.5 - 10, 10, bw + 20, 20);
        ctx.fillStyle = '#ffcf64';
        ctx.fillText(bannerText, cx, 24);

        if (spawnStart2D && spawnCur2D) {
            const sx0 = toScreenX(spawnStart2D.x);
            const sy0 = toScreenY(spawnStart2D.z);
            const sx1 = toScreenX(spawnStart2D.x + (spawnStart2D.x - spawnCur2D.x));
            const sy1 = toScreenY(spawnStart2D.z + (spawnStart2D.z - spawnCur2D.z));

            // Origin ghost circle
            ctx.strokeStyle = 'rgba(255, 180, 40, 0.95)';
            ctx.lineWidth = 2;
            ctx.beginPath();
            ctx.arc(sx0, sy0, 8, 0, Math.PI * 2);
            ctx.stroke();
            ctx.fillStyle = '#ffcc00';
            ctx.beginPath();
            ctx.arc(sx0, sy0, 3, 0, Math.PI * 2);
            ctx.fill();

            // Slingshot pull line
            ctx.strokeStyle = 'rgba(50, 255, 150, 0.9)';
            ctx.lineWidth = 2.5;
            ctx.beginPath();
            ctx.moveTo(sx0, sy0);
            ctx.lineTo(sx1, sy1);
            ctx.stroke();

            const vDist = Math.hypot(spawnStart2D.x - spawnCur2D.x, spawnStart2D.z - spawnCur2D.z) * 0.25;
            ctx.fillStyle = '#78ffb4';
            ctx.font = 'bold 10px sans-serif';
            ctx.textAlign = 'left';
            ctx.fillText(`|V| = ${vDist.toFixed(2)}`, sx1 + 8, sy1);
        }
    }
}

// Start Initialization
window.addEventListener('DOMContentLoaded', initWasm);
