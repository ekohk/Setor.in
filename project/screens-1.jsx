/* global React, Icon, MATERIALS, formatRp */

// ───────────────────────────────────────── HOME
function HomeScreen({ go, balance, t }) {
  const prices = MATERIALS.slice(0, 6).map((m, i) => ({
    ...m,
    trend: i % 3 === 0 ? 'down' : 'up',
    delta: (1 + (i * 1.7) % 4).toFixed(1),
  }));

  return (
    <div className="screen page-enter">
      <div className="screen-bg" />
      <div className="scroll">
        <div className="top-header">
          <div style={{display:'flex',alignItems:'center',gap:12}}>
            <div className="avatar">AR</div>
            <div>
              <div className="greet">Good morning 🌱</div>
              <div className="name">Aria Putri</div>
            </div>
          </div>
          <div className="icon-btn" onClick={() => go('notif')}>
            <Icon name="bell" size={20}/>
            <span className="dot"/>
          </div>
        </div>

        <div className="wallet-card">
          <div className="label">EcoCycle Wallet</div>
          <div className="amount"><span className="currency">Rp</span>{Math.round(balance).toLocaleString('id-ID')}</div>
          <div className="stats">
            <div className="stat-item"><Icon name="leaf" size={14}/><span><strong>42.6</strong> kg recycled</span></div>
            <div className="stat-item"><Icon name="trend-up" size={14}/><span><strong>+18%</strong> this month</span></div>
          </div>
          <svg className="leaf-bg" viewBox="0 0 100 100" fill="#fff" style={{position:'absolute', right:-20, bottom:-30, width:160, height:160, opacity:0.08, pointerEvents:'none'}}>
            <path d="M85 15c0 35-20 55-50 55-8 0-15-3-15-3s15-3 25-18c-8 4-18 4-18 4s8-12 22-18c-12 0-18 4-18 4s15-20 54-24z"/>
          </svg>
        </div>

        <div className="quick-actions">
          <button className="qa-btn primary" onClick={() => go('sell')}>
            <div className="qa-icon"><Icon name="plus" size={20}/></div>
            <div>
              <div className="qa-title">Sell Now</div>
              <div className="qa-sub">Earn from your scraps</div>
            </div>
          </button>
          <button className="qa-btn" onClick={() => go('map')}>
            <div className="qa-icon" style={{background:'oklch(0.95 0.05 230)', color:'oklch(0.5 0.14 230)'}}><Icon name="pin" size={18}/></div>
            <div>
              <div className="qa-title">Find Collectors</div>
              <div className="qa-sub">12 nearby</div>
            </div>
          </button>
        </div>

        <div className="section-head">
          <h3>Today's prices</h3>
          <span className="more" onClick={() => go('prices')}>See all</span>
        </div>

        <div className="price-list">
          {prices.map(p => (
            <div className="price-row" key={p.id}>
              <div className="price-icon"><MaterialIcon name={p.id} size={18}/></div>
              <div className="price-info">
                <div className="name">{p.name}</div>
                <div className="meta">per kilogram</div>
              </div>
              <div className="price-value">
                <div className="amount">{formatRp(p.unit)}</div>
                <div className={`trend ${p.trend}`}>
                  {p.trend === 'up' ? '↗' : '↘'} {p.delta}%
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="banner">
          <div className="leaf">
            <Icon name="leaf-fill" size={22}/>
          </div>
          <div className="text">
            <div className="title">Earn 2× points this week</div>
            <div className="sub">Recycle 5kg+ of plastic and unlock a bonus.</div>
          </div>
          <Icon name="chevron-right" size={18}/>
        </div>
      </div>
    </div>
  );
}

// ───────────────────────────────────────── SELL
function SellScreen({ go, t, onSubmit }) {
  const [material, setMaterial] = React.useState('plastic');
  const [weight, setWeight] = React.useState(3.5);
  const [photos, setPhotos] = React.useState([true, true, false, false]);
  const [method, setMethod] = React.useState('pickup');
  const sel = MATERIALS.find(m => m.id === material);
  const estimate = sel ? sel.unit * weight : 0;

  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div className="icon-btn" onClick={() => go('home')}><Icon name="arrow-left" size={20}/></div>
          <div style={{fontSize:13, fontWeight:600}}>Sell scrap</div>
          <div style={{width:42}}/>
        </div>

        <div className="page-title">What are you selling?</div>
        <div className="page-sub">Pick a material and we'll do the math.</div>

        <div className="field-group">
          <div className="field-label">Material</div>
          <div className="scrap-grid">
            {MATERIALS.slice(0, 6).map(m => (
              <div key={m.id}
                className={`scrap-tile ${material === m.id ? 'active' : ''}`}
                onClick={() => setMaterial(m.id)}>
                <div className="ico"><MaterialIcon name={m.id} size={20}/></div>
                <div className="lbl">{m.name}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="field-group">
          <div className="field-label">Estimated weight</div>
          <div className="weight-input">
            <input type="number" value={weight} step="0.5" min="0"
              onChange={e => setWeight(parseFloat(e.target.value || 0))}/>
            <span className="unit">kg</span>
          </div>
          <div className="weight-stepper">
            {[1, 2.5, 5, 10].map(v => (
              <div key={v} className={`chip ${weight === v ? 'active' : ''}`} onClick={() => setWeight(v)}>{v} kg</div>
            ))}
          </div>
          <div className="weight-est">
            <span>Estimated payout</span>
            <strong>{formatRp(estimate)}</strong>
          </div>
        </div>

        <div className="field-group">
          <div className="field-label">Photos ({photos.filter(Boolean).length}/4)</div>
          <div className="upload-row">
            {photos.map((p, i) => (
              <div key={i}
                className={`upload-tile ${p ? 'filled' : ''}`}
                style={p ? {background: `linear-gradient(135deg, oklch(0.78 0.14 ${120 + i*30}), oklch(0.62 0.16 ${100 + i*40}))`} : {}}
                onClick={() => {
                  const np = [...photos]; np[i] = !np[i]; setPhotos(np);
                }}>
                {!p && <Icon name="camera" size={22}/>}
              </div>
            ))}
          </div>
          <div style={{fontSize:11, color:'var(--ink-3)', marginTop:8}}>
            Add clear photos so the collector can verify quickly.
          </div>
        </div>

        <div className="field-group">
          <div className="field-label">How will you hand it over?</div>
          <div className="method-row">
            <div className={`method-card ${method === 'pickup' ? 'active' : ''}`} onClick={() => setMethod('pickup')}>
              <div className="icon-wrap"><Icon name="home-pickup" size={20}/></div>
              <div className="m-title">Pickup at home</div>
              <div className="m-sub">Free over 5 kg</div>
              <div className="check"><Icon name="check" size={12} strokeWidth={3} stroke="#fff"/></div>
            </div>
            <div className={`method-card ${method === 'dropoff' ? 'active' : ''}`} onClick={() => setMethod('dropoff')}>
              <div className="icon-wrap"><Icon name="box" size={20}/></div>
              <div className="m-title">Drop-off</div>
              <div className="m-sub">+5% bonus</div>
              <div className="check"><Icon name="check" size={12} strokeWidth={3} stroke="#fff"/></div>
            </div>
          </div>
        </div>

        <button className="cta" onClick={() => onSubmit && onSubmit({material, weight, method, estimate})}>
          <Icon name="check" size={18} strokeWidth={2.5}/>
          Submit & track order
        </button>
      </div>
    </div>
  );
}

// ───────────────────────────────────────── MAP
function MapScreen({ go, t }) {
  const [filter, setFilter] = React.useState('all');
  const [selected, setSelected] = React.useState('budi');

  const collectors = [
    { id: 'budi', name: 'Budi Recycling', x: 35, y: 32, dist: '0.8 km', rating: 4.9, jobs: 230, eta: '15 min' },
    { id: 'sari', name: 'Sari Daur Ulang', x: 68, y: 22, dist: '1.4 km', rating: 4.7, jobs: 142, eta: '22 min' },
    { id: 'tani', name: 'Eco Tani', x: 22, y: 58, dist: '2.1 km', rating: 4.8, jobs: 305, eta: '28 min' },
    { id: 'agus', name: 'Agus Logam', x: 75, y: 60, dist: '2.6 km', rating: 4.6, jobs: 88, eta: '32 min' },
  ];

  const sel = collectors.find(c => c.id === selected);

  return (
    <div className="screen page-enter">
      <div className="map-canvas">
        <svg className="map-svg" viewBox="0 0 100 100" preserveAspectRatio="none">
          <defs>
            <pattern id="grid" width="8" height="8" patternUnits="userSpaceOnUse">
              <path d="M 8 0 L 0 0 0 8" fill="none" stroke="rgba(20,80,40,0.05)" strokeWidth="0.2"/>
            </pattern>
          </defs>
          <rect width="100" height="100" fill="url(#grid)"/>
          {/* abstract roads */}
          <path d="M -5 30 Q 30 28 55 40 T 105 50" fill="none" stroke="rgba(255,255,255,0.7)" strokeWidth="2.5"/>
          <path d="M -5 30 Q 30 28 55 40 T 105 50" fill="none" stroke="rgba(20,80,40,0.12)" strokeWidth="2.5" strokeDasharray="0.5 1"/>
          <path d="M 50 -5 Q 48 25 60 50 T 70 105" fill="none" stroke="rgba(255,255,255,0.7)" strokeWidth="2"/>
          <path d="M -5 75 Q 25 70 50 78 T 105 80" fill="none" stroke="rgba(255,255,255,0.7)" strokeWidth="2"/>
          {/* parks / blocks */}
          <rect x="10" y="40" width="22" height="14" rx="2" fill="oklch(0.88 0.10 145 / 0.5)"/>
          <rect x="60" y="65" width="20" height="18" rx="2" fill="oklch(0.88 0.10 145 / 0.5)"/>
          <rect x="38" y="10" width="14" height="12" rx="2" fill="oklch(0.92 0.06 80 / 0.5)"/>
          <rect x="78" y="38" width="16" height="10" rx="2" fill="oklch(0.92 0.06 80 / 0.5)"/>
          {/* user location */}
          <circle cx="50" cy="50" r="2.5" fill="#0ea5e9"/>
          <circle cx="50" cy="50" r="6" fill="rgba(14,165,233,0.2)"/>
        </svg>

        {collectors.map(c => (
          <div key={c.id} className={`map-marker ${selected === c.id ? 'active' : ''}`}
            style={{left: `${c.x}%`, top: `${c.y}%`}}
            onClick={() => setSelected(c.id)}>
            <div className="pin">
              <span className="dot"/>
              <span>{c.name.split(' ')[0]}</span>
            </div>
            {selected !== c.id && <div className="pulse"/>}
          </div>
        ))}
      </div>

      <div className="map-search">
        <Icon name="search" size={18}/>
        <input placeholder="Search area or collector..."/>
        <Icon name="filter" size={18}/>
      </div>

      <div className="map-filters">
        {[
          {id:'all', label:'All', icon: 'pin'},
          {id:'near', label:'Nearest', icon: 'navigation'},
          {id:'top', label:'Top rated', icon: 'star'},
          {id:'plastic', label:'Plastic'},
          {id:'metal', label:'Metal'},
        ].map(f => (
          <div key={f.id} className={`map-filter-chip ${filter === f.id ? 'active' : ''}`} onClick={() => setFilter(f.id)}>
            {f.icon && <Icon name={f.icon} size={13}/>}
            {f.label}
          </div>
        ))}
      </div>

      {sel && (
        <div className="collector-card">
          <div className="collector-head">
            <div className="collector-avatar">{sel.name.split(' ').map(s=>s[0]).slice(0,2).join('')}</div>
            <div className="collector-info">
              <div className="name">{sel.name}</div>
              <div className="meta">
                <span style={{display:'inline-flex',alignItems:'center',gap:3}}>
                  <Icon name="star" size={12} stroke="oklch(0.7 0.16 80)"/>
                  {sel.rating}
                </span>
                <span className="dot-sep"/>
                <span>{sel.dist}</span>
                <span className="dot-sep"/>
                <span>~{sel.eta}</span>
              </div>
            </div>
            <div style={{padding:'4px 10px', borderRadius:999, background:'var(--accent-soft)', color:'var(--accent-deep)', fontSize:11, fontWeight:700}}>Open</div>
          </div>
          <div className="collector-stats">
            <div className="collector-stat">
              <div className="v">{sel.jobs}</div>
              <div className="l">Pickups done</div>
            </div>
            <div className="collector-stat">
              <div className="v">All</div>
              <div className="l">Materials</div>
            </div>
            <div className="collector-stat">
              <div className="v">07–21</div>
              <div className="l">Open hours</div>
            </div>
          </div>
          <div className="collector-cta">
            <button className="btn-secondary"><Icon name="navigation" size={14}/> Directions</button>
            <button className="btn-primary" onClick={() => go('sell')}><Icon name="plus" size={14} strokeWidth={2.5}/> Sell to {sel.name.split(' ')[0]}</button>
          </div>
        </div>
      )}
    </div>
  );
}

window.HomeScreen = HomeScreen;
window.SellScreen = SellScreen;
window.MapScreen = MapScreen;
