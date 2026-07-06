'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import * as THREE from 'three';
import { StockMention } from '@/lib/stockExtractor';
import { useNewsStore } from '@/store/newsStore';

interface HistoryPoint {
  t: number;
  price: number;
}

interface SymbolHistory {
  symbol: string;
  points: HistoryPoint[];
  publishPrice: number | null;
  changePercent: number | null;
}

interface MarketReaction3DProps {
  isOpen: boolean;
  onClose: () => void;
  articleTitle: string;
  publishedAt: string;
  mentions: StockMention[];
}

// Chart palette (validated: see dataviz status pair, both surfaces)
const PALETTE = {
  light: {
    surface: '#fcfcfb',
    grid: '#e1e0d9',
    baseline: '#c3c2b7',
    muted: '#898781',
    ink: '#0b0b0b',
    up: '#0ca30c',
    down: '#d03b3b',
    upText: '#006300',
    downText: '#d03b3b',
  },
  dark: {
    surface: '#1a1a19',
    grid: '#2c2c2a',
    baseline: '#383835',
    muted: '#898781',
    ink: '#ffffff',
    up: '#0ca30c',
    down: '#d03b3b',
    upText: '#0ca30c',
    downText: '#e66767',
  },
};

// Scene layout constants
const X_SPAN = 10; // time axis width
const Y_SPAN = 2.2; // height for the largest |% change|
const ROW_GAP = 1.7; // z distance between symbol rows
const RIBBON_DEPTH = 0.7;

function makeTextSprite(text: string, color: string, bold = false): THREE.Sprite {
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d')!;
  const font = `${bold ? '700' : '500'} 44px system-ui, -apple-system, "Segoe UI", sans-serif`;
  ctx.font = font;
  const w = Math.ceil(ctx.measureText(text).width) + 16;
  canvas.width = w;
  canvas.height = 64;
  ctx.font = font;
  ctx.fillStyle = color;
  ctx.textBaseline = 'middle';
  ctx.fillText(text, 8, 32);

  const texture = new THREE.CanvasTexture(canvas);
  texture.colorSpace = THREE.SRGBColorSpace;
  const sprite = new THREE.Sprite(
    new THREE.SpriteMaterial({ map: texture, transparent: true, depthTest: false })
  );
  const scale = 0.0105;
  sprite.scale.set(w * scale, 64 * scale, 1);
  return sprite;
}

export default function MarketReaction3D({
  isOpen,
  onClose,
  articleTitle,
  publishedAt,
  mentions,
}: MarketReaction3DProps) {
  const isDarkMode = useNewsStore((state) => state.isDarkMode);
  const containerRef = useRef<HTMLDivElement>(null);
  const [histories, setHistories] = useState<SymbolHistory[] | null>(null);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [view, setView] = useState<'3d' | 'table'>('3d');
  const [hover, setHover] = useState<{
    left: number;
    top: number;
    symbol: string;
    time: string;
    price: number;
    pct: number;
  } | null>(null);

  const symbols = mentions
    .filter((m) => m.type === 'stock' || m.type === 'etf')
    .map((m) => m.symbol)
    .slice(0, 4);

  // Fetch the 24h-after-publish price history
  useEffect(() => {
    if (!isOpen || symbols.length === 0) return;
    let cancelled = false;

    const load = async () => {
      try {
        const response = await fetch(
          `/api/stock/history?symbols=${encodeURIComponent(symbols.join(','))}&from=${encodeURIComponent(publishedAt)}`
        );
        const data = await response.json();
        if (!response.ok) throw new Error(data.message || 'Failed to load market data');
        if (!cancelled) setHistories(data.histories);
      } catch (err) {
        if (!cancelled) setFetchError(err instanceof Error ? err.message : 'Failed to load market data');
      }
    };

    load();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, publishedAt, symbols.join(',')]);

  // Build and drive the three.js scene
  useEffect(() => {
    const container = containerRef.current;
    if (!container || view !== '3d' || !histories) return;

    const colors = isDarkMode ? PALETTE.dark : PALETTE.light;
    const drawable = histories.filter((h) => h.points.length >= 2 && h.publishPrice);
    if (drawable.length === 0) return;

    // ----- scales -----
    const t0 = Math.min(...drawable.map((h) => h.points[0].t));
    const t1 = Math.max(...drawable.map((h) => h.points[h.points.length - 1].t));
    const tRange = Math.max(t1 - t0, 1);
    // shared y scale: % change since publish, symmetric around 0 (one axis for all rows)
    const maxAbsPct = Math.max(
      0.5,
      ...drawable.flatMap((h) =>
        h.points.map((p) => Math.abs(((p.price - h.publishPrice!) / h.publishPrice!) * 100))
      )
    );
    const xOf = (t: number) => ((t - t0) / tRange) * X_SPAN - X_SPAN / 2;
    const yOf = (pct: number) => (pct / maxAbsPct) * Y_SPAN;
    const zOf = (row: number) => (row - (drawable.length - 1) / 2) * ROW_GAP;

    // ----- scene -----
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(colors.surface);
    scene.add(new THREE.AmbientLight(0xffffff, 2.2));
    const dir = new THREE.DirectionalLight(0xffffff, 1.4);
    dir.position.set(4, 8, 6);
    scene.add(dir);

    const width = container.clientWidth;
    const height = container.clientHeight;
    const camera = new THREE.PerspectiveCamera(40, width / height, 0.1, 100);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setSize(width, height);
    container.appendChild(renderer.domElement);

    // recessive grid on the baseline plane
    const gridSize = Math.max(X_SPAN + 2, drawable.length * ROW_GAP + 2);
    const grid = new THREE.GridHelper(gridSize, 12, colors.baseline, colors.grid);
    scene.add(grid);

    const upColor = new THREE.Color(colors.up);
    const downColor = new THREE.Color(colors.down);
    const pickMeshes: THREE.Mesh[] = [];

    for (let row = 0; row < drawable.length; row++) {
      const h = drawable[row];
      const base = h.publishPrice!;
      const z = zOf(row);
      const n = h.points.length;

      const xs: number[] = [];
      const ys: number[] = [];
      for (const p of h.points) {
        xs.push(xOf(p.t));
        ys.push(yOf(((p.price - base) / base) * 100));
      }

      // extruded ribbon: front wall, back wall, top cap — vertex-colored by sign
      const positions: number[] = [];
      const vertexColors: number[] = [];
      const pushQuad = (
        a: THREE.Vector3, b: THREE.Vector3, c: THREE.Vector3, d: THREE.Vector3,
        colA: THREE.Color, colB: THREE.Color
      ) => {
        // two triangles a-b-c, a-c-d; colA on a/d edge, colB on b/c edge
        positions.push(a.x, a.y, a.z, b.x, b.y, b.z, c.x, c.y, c.z);
        vertexColors.push(colA.r, colA.g, colA.b, colB.r, colB.g, colB.b, colB.r, colB.g, colB.b);
        positions.push(a.x, a.y, a.z, c.x, c.y, c.z, d.x, d.y, d.z);
        vertexColors.push(colA.r, colA.g, colA.b, colB.r, colB.g, colB.b, colA.r, colA.g, colA.b);
      };

      const half = RIBBON_DEPTH / 2;
      for (let i = 0; i < n - 1; i++) {
        const cL = ys[i] >= 0 ? upColor : downColor;
        const cR = ys[i + 1] >= 0 ? upColor : downColor;
        // front wall (z + half): baseline -> value
        pushQuad(
          new THREE.Vector3(xs[i], 0, z + half),
          new THREE.Vector3(xs[i + 1], 0, z + half),
          new THREE.Vector3(xs[i + 1], ys[i + 1], z + half),
          new THREE.Vector3(xs[i], ys[i], z + half),
          cL, cR
        );
        // back wall
        pushQuad(
          new THREE.Vector3(xs[i + 1], 0, z - half),
          new THREE.Vector3(xs[i], 0, z - half),
          new THREE.Vector3(xs[i], ys[i], z - half),
          new THREE.Vector3(xs[i + 1], ys[i + 1], z - half),
          cR, cL
        );
        // top cap
        pushQuad(
          new THREE.Vector3(xs[i], ys[i], z + half),
          new THREE.Vector3(xs[i + 1], ys[i + 1], z + half),
          new THREE.Vector3(xs[i + 1], ys[i + 1], z - half),
          new THREE.Vector3(xs[i], ys[i], z - half),
          cL, cR
        );
      }

      const geometry = new THREE.BufferGeometry();
      geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
      geometry.setAttribute('color', new THREE.Float32BufferAttribute(vertexColors, 3));
      geometry.computeVertexNormals();
      const mesh = new THREE.Mesh(
        geometry,
        new THREE.MeshLambertMaterial({
          vertexColors: true,
          transparent: true,
          opacity: 0.42,
          side: THREE.DoubleSide,
        })
      );
      mesh.userData = { history: h, xs, ys, z };
      scene.add(mesh);
      pickMeshes.push(mesh);

      // crest line along the top edge (the actual price line)
      const crestGeometry = new THREE.BufferGeometry().setFromPoints(
        xs.map((x, i) => new THREE.Vector3(x, ys[i], z + half))
      );
      const crestColors: number[] = [];
      for (let i = 0; i < n; i++) {
        const c = ys[i] >= 0 ? upColor : downColor;
        crestColors.push(c.r, c.g, c.b);
      }
      crestGeometry.setAttribute('color', new THREE.Float32BufferAttribute(crestColors, 3));
      scene.add(new THREE.Line(crestGeometry, new THREE.LineBasicMaterial({ vertexColors: true })));

      // direct label: symbol at the end of its ribbon
      const label = makeTextSprite(h.symbol, colors.ink, true);
      label.position.set(xs[n - 1] + 0.55, Math.max(ys[n - 1], 0.18) + 0.2, z);
      scene.add(label);
    }

    // axis labels (muted ink, few and selective)
    const frontZ = zOf(drawable.length - 1) + 0.95;
    const backZ = zOf(0) - 0.9;
    const publishLabel = makeTextSprite('publish', colors.muted);
    publishLabel.position.set(-X_SPAN / 2, 0.16, frontZ);
    scene.add(publishLabel);
    const endLabel = makeTextSprite('+24h', colors.muted);
    endLabel.position.set(X_SPAN / 2 - 0.4, 0.16, frontZ);
    scene.add(endLabel);
    const topLabel = makeTextSprite(`+${maxAbsPct.toFixed(1)}%`, colors.muted);
    topLabel.position.set(-X_SPAN / 2 - 0.7, Y_SPAN, backZ);
    scene.add(topLabel);
    const zeroLabel = makeTextSprite('0%', colors.muted);
    zeroLabel.position.set(-X_SPAN / 2 - 0.7, 0.05, backZ);
    scene.add(zeroLabel);

    // hover marker
    const marker = new THREE.Mesh(
      new THREE.SphereGeometry(0.09, 16, 16),
      new THREE.MeshBasicMaterial({ color: colors.ink })
    );
    marker.visible = false;
    scene.add(marker);

    // ----- hand-rolled orbit (drag to rotate, wheel to zoom) -----
    const target = new THREE.Vector3(0, 0.4, 0);
    let azimuth = 0.45;
    let polar = 1.0;
    let radius = 13;
    let azimuthGoal = azimuth;
    let polarGoal = polar;
    let radiusGoal = radius;
    let dragging = false;
    let lastX = 0;
    let lastY = 0;

    const applyCamera = () => {
      camera.position.set(
        target.x + radius * Math.sin(polar) * Math.sin(azimuth),
        target.y + radius * Math.cos(polar),
        target.z + radius * Math.sin(polar) * Math.cos(azimuth)
      );
      camera.lookAt(target);
    };

    const onPointerDown = (e: PointerEvent) => {
      dragging = true;
      lastX = e.clientX;
      lastY = e.clientY;
      renderer.domElement.setPointerCapture(e.pointerId);
    };
    const onPointerUp = (e: PointerEvent) => {
      dragging = false;
      renderer.domElement.releasePointerCapture(e.pointerId);
    };
    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      radiusGoal = Math.min(20, Math.max(6, radiusGoal + e.deltaY * 0.012));
    };

    // hover picking
    const raycaster = new THREE.Raycaster();
    const pointer = new THREE.Vector2();
    const onPointerMove = (e: PointerEvent) => {
      if (dragging) {
        azimuthGoal += (e.clientX - lastX) * 0.006;
        polarGoal = Math.min(1.45, Math.max(0.2, polarGoal + (e.clientY - lastY) * 0.005));
        lastX = e.clientX;
        lastY = e.clientY;
        return;
      }
      const rect = renderer.domElement.getBoundingClientRect();
      pointer.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
      pointer.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;
      raycaster.setFromCamera(pointer, camera);
      const hit = raycaster.intersectObjects(pickMeshes)[0];
      if (!hit) {
        marker.visible = false;
        setHover(null);
        return;
      }
      const { history, xs, ys, z } = hit.object.userData as {
        history: SymbolHistory;
        xs: number[];
        ys: number[];
        z: number;
      };
      // snap to the nearest data point in time
      let index = 0;
      let best = Infinity;
      for (let i = 0; i < xs.length; i++) {
        const d = Math.abs(xs[i] - hit.point.x);
        if (d < best) {
          best = d;
          index = i;
        }
      }
      const point = history.points[index];
      const pct = ((point.price - history.publishPrice!) / history.publishPrice!) * 100;
      marker.position.set(xs[index], ys[index], z + RIBBON_DEPTH / 2);
      marker.visible = true;
      setHover({
        left: e.clientX - rect.left,
        top: e.clientY - rect.top,
        symbol: history.symbol,
        time: new Date(point.t).toLocaleString([], {
          weekday: 'short',
          hour: '2-digit',
          minute: '2-digit',
        }),
        price: point.price,
        pct,
      });
    };
    const onPointerLeave = () => {
      marker.visible = false;
      setHover(null);
    };

    renderer.domElement.addEventListener('pointerdown', onPointerDown);
    renderer.domElement.addEventListener('pointerup', onPointerUp);
    renderer.domElement.addEventListener('pointermove', onPointerMove);
    renderer.domElement.addEventListener('pointerleave', onPointerLeave);
    renderer.domElement.addEventListener('wheel', onWheel, { passive: false });
    renderer.domElement.style.cursor = 'grab';
    renderer.domElement.style.touchAction = 'none';

    let frame = 0;
    const animate = () => {
      frame = requestAnimationFrame(animate);
      // damped approach to the goal angles
      azimuth += (azimuthGoal - azimuth) * 0.12;
      polar += (polarGoal - polar) * 0.12;
      radius += (radiusGoal - radius) * 0.12;
      applyCamera();
      renderer.render(scene, camera);
    };
    animate();

    const resizeObserver = new ResizeObserver(() => {
      const w = container.clientWidth;
      const h = container.clientHeight;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    });
    resizeObserver.observe(container);

    return () => {
      cancelAnimationFrame(frame);
      resizeObserver.disconnect();
      renderer.domElement.removeEventListener('pointerdown', onPointerDown);
      renderer.domElement.removeEventListener('pointerup', onPointerUp);
      renderer.domElement.removeEventListener('pointermove', onPointerMove);
      renderer.domElement.removeEventListener('pointerleave', onPointerLeave);
      renderer.domElement.removeEventListener('wheel', onWheel);
      scene.traverse((obj) => {
        if (obj instanceof THREE.Mesh || obj instanceof THREE.Line || obj instanceof THREE.Sprite) {
          obj.geometry?.dispose();
          const material = (obj as THREE.Mesh).material as THREE.Material;
          material?.dispose();
        }
      });
      renderer.dispose();
      container.removeChild(renderer.domElement);
    };
  }, [histories, isDarkMode, view]);

  const formatPct = useCallback(
    (pct: number) => `${pct >= 0 ? '+' : ''}${pct.toFixed(2)}%`,
    []
  );

  if (!isOpen) return null;

  const colors = isDarkMode ? PALETTE.dark : PALETTE.light;
  const drawable = histories?.filter((h) => h.points.length >= 2 && h.publishPrice) ?? [];
  const publishDate = new Date(publishedAt);

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity" onClick={onClose} />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-4xl overflow-hidden">
          {/* Header */}
          <div className="flex items-start justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <div className="min-w-0">
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
                24h Market Reaction
              </h2>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">
                {articleTitle}
              </p>
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
                24 hours after publication · {publishDate.toLocaleString([], {
                  dateStyle: 'medium',
                  timeStyle: 'short',
                })}
              </p>
            </div>
            <div className="flex items-center gap-2 flex-shrink-0 ml-4">
              <div className="flex rounded-lg border border-gray-200 dark:border-gray-600 overflow-hidden text-sm">
                <button
                  onClick={() => setView('3d')}
                  className={`px-3 py-1.5 font-medium transition-colors ${
                    view === '3d'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  3D
                </button>
                <button
                  onClick={() => setView('table')}
                  className={`px-3 py-1.5 font-medium transition-colors ${
                    view === 'table'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  Table
                </button>
              </div>
              <button
                onClick={onClose}
                className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors duration-200"
              >
                <svg className="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          {/* Legend / summary chips */}
          {histories && (
            <div className="flex flex-wrap gap-2 px-4 pt-3">
              {histories.map((h) => (
                <div
                  key={h.symbol}
                  className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-gray-100 dark:bg-gray-700 text-sm"
                >
                  <span className="font-bold text-gray-900 dark:text-gray-100">{h.symbol}</span>
                  {h.changePercent != null ? (
                    <>
                      <span
                        aria-hidden
                        style={{ color: h.changePercent >= 0 ? colors.upText : colors.downText }}
                      >
                        {h.changePercent >= 0 ? '▲' : '▼'}
                      </span>
                      <span className="text-gray-700 dark:text-gray-200">
                        {formatPct(h.changePercent)}
                        <span className="sr-only">{h.changePercent >= 0 ? ' up' : ' down'} since publish</span>
                      </span>
                    </>
                  ) : (
                    <span className="text-gray-500 dark:text-gray-400 text-xs">no trading data</span>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Body */}
          <div className="p-4">
            {fetchError ? (
              <div className="py-16 text-center text-red-600 dark:text-red-400 text-sm">{fetchError}</div>
            ) : !histories ? (
              <div className="py-16 flex items-center justify-center">
                <svg className="animate-spin h-8 w-8 text-blue-600" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              </div>
            ) : drawable.length === 0 ? (
              <div className="py-16 text-center text-gray-500 dark:text-gray-400 text-sm">
                The market was closed during the 24 hours after publication — no trading data for{' '}
                {symbols.join(', ')}.
              </div>
            ) : view === '3d' ? (
              <div className="relative">
                <div
                  ref={containerRef}
                  className="w-full h-[420px] rounded-lg overflow-hidden"
                  style={{ backgroundColor: colors.surface }}
                />
                {hover && (
                  <div
                    className="pointer-events-none absolute z-10 px-2.5 py-1.5 rounded-lg shadow-lg text-xs
                               bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-600"
                    style={{ left: Math.min(hover.left + 14, 720), top: Math.max(hover.top - 44, 4) }}
                  >
                    <span className="font-bold text-gray-900 dark:text-gray-100">{hover.symbol}</span>
                    <span className="text-gray-500 dark:text-gray-400"> · {hover.time} · </span>
                    <span className="text-gray-900 dark:text-gray-100">${hover.price.toFixed(2)}</span>
                    <span className="text-gray-500 dark:text-gray-400"> · </span>
                    <span style={{ color: hover.pct >= 0 ? colors.upText : colors.downText }}>
                      {hover.pct >= 0 ? '▲' : '▼'}
                    </span>{' '}
                    <span className="text-gray-900 dark:text-gray-100">{formatPct(hover.pct)}</span>
                  </div>
                )}
                <p className="mt-2 text-xs text-gray-400 dark:text-gray-500 text-center">
                  Drag to rotate · scroll to zoom · hover a ribbon for details. Height = % change since publish.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
                      <th className="py-2 pr-4">Symbol</th>
                      <th className="py-2 pr-4">At publish</th>
                      <th className="py-2 pr-4">High</th>
                      <th className="py-2 pr-4">Low</th>
                      <th className="py-2 pr-4">Last</th>
                      <th className="py-2">Change</th>
                    </tr>
                  </thead>
                  <tbody>
                    {drawable.map((h) => {
                      const prices = h.points.map((p) => p.price);
                      const last = prices[prices.length - 1];
                      return (
                        <tr key={h.symbol} className="border-b border-gray-100 dark:border-gray-700/50">
                          <td className="py-2 pr-4 font-bold text-gray-900 dark:text-gray-100">{h.symbol}</td>
                          <td className="py-2 pr-4 text-gray-700 dark:text-gray-300">${h.publishPrice!.toFixed(2)}</td>
                          <td className="py-2 pr-4 text-gray-700 dark:text-gray-300">${Math.max(...prices).toFixed(2)}</td>
                          <td className="py-2 pr-4 text-gray-700 dark:text-gray-300">${Math.min(...prices).toFixed(2)}</td>
                          <td className="py-2 pr-4 text-gray-700 dark:text-gray-300">${last.toFixed(2)}</td>
                          <td className="py-2 text-gray-900 dark:text-gray-100">
                            <span
                              aria-hidden
                              style={{ color: (h.changePercent ?? 0) >= 0 ? colors.upText : colors.downText }}
                            >
                              {(h.changePercent ?? 0) >= 0 ? '▲' : '▼'}
                            </span>{' '}
                            {formatPct(h.changePercent ?? 0)}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
