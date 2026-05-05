/* global React, Icon, MATERIALS, formatRp */

// ───────────────────────────────────────── TRACKING
function TrackingScreen({ go, t }) {
  // Stage selector — let the user explore each phase of the journey
  const [stage, setStage] = React.useState(3); // 0..7

  const steps = [
    { id: 'received', icon: 'box',         label: 'Order received',     time: '09:42', detail: 'We notified Budi Recycling about your 3.5 kg plastic order.' },
    { id: 'accepted', icon: 'check-circle',label: 'Collector accepted', time: '09:48', detail: 'Budi confirmed and assigned driver Wahyu.' },
    { id: 'enroute',  icon: 'truck',       label: 'Driver en-route',    time: '10:02', detail: 'Wahyu is 1.2 km away — ETA 6 min.' },
    { id: 'arrived',  icon: 'pin',         label: 'Driver arrived',     time: '10:15', detail: 'Wahyu is at your gate. Hand over the bag.' },
    { id: 'weighing', icon: 'scale',       label: 'Weighing in progress', time: '10:18', detail: 'Live weight reading from certified scale.' },
    { id: 'quality',  icon: 'shield',      label: 'Quality check',      time: '10:22', detail: 'Verifying material grade.' },
    { id: 'payout',   icon: 'wallet',      label: 'Payout processed',   time: '10:24', detail: 'Funds transferred to your EcoCycle wallet.' },
    { id: 'done',     icon: 'check-circle',label: 'Completed',          time: '10:25', detail: 'Order closed. Thanks for recycling!' },
  ];

  const cur = steps[stage];

  // Live weight animation while on weighing stage
  const [liveWeight, setLiveWeight] = React.useState(0);
  React.useEffect(() => {
    if (stage !== 4) { setLiveWeight(stage > 4 ? 3.7 : 0); return; }
    let v = 0; setLiveWeight(0);
    const id = setInterval(() => {
      v += 0.08 + Math.random() * 0.18;
      if (v >= 3.7) { v = 3.7; clearInterval(id); }
      setLiveWeight(v);
    }, 80);
    return () => clearInterval(id);
  }, [stage]);

  const estWeight = 3.5;
  const actualWeight = stage >= 4 ? (stage === 4 ? liveWeight : 3.7) : null;
  const unitPrice = 5500;
  const estPayout = estWeight * unitPrice;
  const actualPayout = actualWeight ? actualWeight * unitPrice : null;

  // Progress percentage for hero ring
  const pct = Math.round(((stage + 1) / steps.length) * 100);

  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div className="icon-btn" onClick={() => go('home')}><Icon name="arrow-left" size={20}/></div>
          <div style={{fontSize:13, fontWeight:600}}>Order tracking</div>
          <div className="icon-btn"><Icon name="message" size={18}/></div>
        </div>

        {/* Hero status card */}
        <div className="track-hero">
          <div className="track-hero-top">
            <div>
              <div className="order-id">ORDER · ECC-04827</div>
              <div className="track-status-title">{cur.label}</div>
              <div className="track-status-sub">{cur.detail}</div>
            </div>
            <ProgressRing pct={pct}/>
          </div>

          {/* Stage scrubber — tap to preview each phase */}
          <div className="stage-scrubber">
            {steps.map((s, i) => (
              <button key={s.id}
                className={`scrub-dot ${i < stage ? 'done' : i === stage ? 'current' : ''}`}
                onClick={() => setStage(i)}
                title={s.label}>
                <span className="scrub-fill"/>
              </button>
            ))}
          </div>
          <div className="stage-scrubber-meta">
            <span>Step {stage + 1} of {steps.length}</span>
            <span>Tap a dot to preview</span>
          </div>
        </div>

        {/* CONTEXTUAL CARD — content adapts to stage */}
        {stage === 2 && <DriverEnRouteCard/>}
        {stage === 3 && <ArrivedCard/>}
        {stage === 4 && <WeighingCard live={liveWeight} estimate={estWeight}/>}
        {stage === 5 && <QualityCard/>}
        {(stage === 6 || stage === 7) && <PayoutCard actualWeight={actualWeight} unitPrice={unitPrice} done={stage === 7}/>}
        {stage < 2 && <PreparingCard step={stage}/>}

        {/* Order summary — always visible */}
        <div className="section-head"><h3>Order details</h3></div>
        <div className="tracker-card">
          <div className="ord-row">
            <div className="ord-row-l">Material</div>
            <div className="ord-row-r">Plastic · PET clear</div>
          </div>
          <div className="ord-row">
            <div className="ord-row-l">Estimated</div>
            <div className="ord-row-r">{estWeight} kg · {formatRp(estPayout)}</div>
          </div>
          <div className="ord-row">
            <div className="ord-row-l">Actual weight</div>
            <div className="ord-row-r">
              {actualWeight ? (
                <>
                  <strong>{actualWeight.toFixed(2)} kg</strong>
                  <span className={`diff ${actualWeight >= estWeight ? 'pos' : 'neg'}`}>
                    {actualWeight >= estWeight ? '+' : ''}{(((actualWeight - estWeight) / estWeight) * 100).toFixed(1)}%
                  </span>
                </>
              ) : <span style={{color:'var(--ink-4)'}}>Pending</span>}
            </div>
          </div>
          <div className="ord-row">
            <div className="ord-row-l">Method</div>
            <div className="ord-row-r">Home pickup · Free</div>
          </div>
          <div className="ord-row" style={{borderTop:'1px solid var(--line)', paddingTop:14, marginTop:4}}>
            <div className="ord-row-l" style={{fontWeight:600, color:'var(--ink)'}}>Final payout</div>
            <div className="ord-row-r" style={{fontSize:16, fontWeight:700}}>
              {actualPayout ? formatRp(actualPayout) : formatRp(estPayout)}
            </div>
          </div>
        </div>

        {/* Driver / collector contact */}
        <div className="tracker-card">
          <div className="collector-mini" style={{margin:0, background:'transparent', padding:0}}>
            <div className="collector-avatar" style={{width:44, height:44, borderRadius:12, fontSize:14}}>WB</div>
            <div className="collector-info" style={{flex:1}}>
              <div className="name" style={{fontSize:14}}>Wahyu</div>
              <div className="meta" style={{fontSize:11}}>
                <span style={{display:'inline-flex',alignItems:'center',gap:3}}>
                  <Icon name="star" size={11}/>4.9
                </span>
                <span className="dot-sep"/>
                <span>Driver · Budi Recycling</span>
              </div>
            </div>
            <div className="cm-actions">
              <div className="cm-icon outline"><Icon name="message" size={16}/></div>
              <div className="cm-icon"><Icon name="phone" size={16}/></div>
            </div>
          </div>

          {/* Last message preview */}
          <div className="chat-preview">
            <div className="chat-bubble">
              "Hi, saya sudah di gerbang depan ya 🙏"
            </div>
            <div className="chat-time">Wahyu · 2 min ago</div>
          </div>
        </div>

        {/* Full timeline */}
        <div className="section-head"><h3>Activity</h3></div>
        <div className="tracker-card">
          <div className="timeline">
            {steps.map((s, i) => {
              const cls = i < stage ? 'done' : i === stage ? 'current' : 'pending';
              return (
                <div className={`tl-step ${cls}`} key={s.id} onClick={() => setStage(i)} style={{cursor:'pointer'}}>
                  <div className="marker">
                    {i < stage && <Icon name="check" size={9} strokeWidth={3} stroke="#fff"/>}
                  </div>
                  <div className="lbl">{s.label}</div>
                  <div className="time">{s.time}</div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

function ProgressRing({ pct }) {
  const r = 26, c = 2 * Math.PI * r;
  return (
    <div className="ring-wrap">
      <svg width="64" height="64" viewBox="0 0 64 64">
        <circle cx="32" cy="32" r={r} fill="none" stroke="rgba(255,255,255,0.15)" strokeWidth="4"/>
        <circle cx="32" cy="32" r={r} fill="none" stroke="#fff" strokeWidth="4"
          strokeDasharray={c} strokeDashoffset={c - (c * pct / 100)}
          strokeLinecap="round" transform="rotate(-90 32 32)"
          style={{transition:'stroke-dashoffset .4s ease'}}/>
      </svg>
      <div className="ring-pct">{pct}%</div>
    </div>
  );
}

function PreparingCard({ step }) {
  return (
    <div className="tracker-card stage-card">
      <div className="stage-card-icon"><Icon name="clock" size={20}/></div>
      <div className="stage-card-title">{step === 0 ? 'Looking for collector' : 'Collector preparing'}</div>
      <div className="stage-card-body">
        {step === 0
          ? 'We\'re notifying nearby certified collectors. This usually takes under 3 minutes.'
          : 'Budi Recycling has accepted your order. Driver will leave shortly.'}
      </div>
    </div>
  );
}

function DriverEnRouteCard() {
  return (
    <div className="tracker-card stage-card no-pad">
      <div className="mini-map">
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" style={{width:'100%',height:'100%'}}>
          <defs>
            <pattern id="g2" width="6" height="6" patternUnits="userSpaceOnUse">
              <path d="M 6 0 L 0 0 0 6" fill="none" stroke="rgba(20,80,40,0.06)" strokeWidth="0.2"/>
            </pattern>
          </defs>
          <rect width="100" height="60" fill="url(#g2)"/>
          <path d="M -5 40 Q 25 38 50 30 T 105 22" fill="none" stroke="rgba(255,255,255,0.9)" strokeWidth="2.5"/>
          <path d="M -5 40 Q 25 38 50 30 T 105 22" fill="none" stroke="var(--accent)" strokeWidth="1.5" strokeDasharray="2 2"/>
          {/* Driver */}
          <circle cx="32" cy="36" r="3" fill="var(--accent)"/>
          <circle cx="32" cy="36" r="6" fill="rgba(47,125,82,0.2)"/>
          {/* Destination */}
          <circle cx="78" cy="24" r="2" fill="var(--ink)"/>
          <rect x="74" y="14" width="8" height="8" rx="1" fill="var(--ink)"/>
        </svg>
        <div className="map-label driver-lbl">
          <Icon name="truck" size={12}/> Wahyu
        </div>
        <div className="map-label dest-lbl">
          <Icon name="pin" size={12}/> You
        </div>
      </div>
      <div style={{padding:'14px 18px 16px'}}>
        <div className="enroute-row">
          <div>
            <div style={{fontSize:11, color:'var(--ink-3)', fontWeight:600, letterSpacing:'0.06em', textTransform:'uppercase'}}>ETA</div>
            <div style={{fontSize:22, fontWeight:700, letterSpacing:'-0.02em', marginTop:2}}>6 min</div>
          </div>
          <div style={{textAlign:'right'}}>
            <div style={{fontSize:11, color:'var(--ink-3)', fontWeight:600, letterSpacing:'0.06em', textTransform:'uppercase'}}>Distance</div>
            <div style={{fontSize:22, fontWeight:700, letterSpacing:'-0.02em', marginTop:2}}>1.2 km</div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ArrivedCard() {
  return (
    <div className="tracker-card stage-card">
      <div className="stage-card-icon arrived-icon"><Icon name="pin" size={20}/></div>
      <div className="stage-card-title">Driver is at your gate</div>
      <div className="stage-card-body">
        Wahyu is waiting at the front. Hand over your scrap bag — please confirm with the verification code below.
      </div>
      <div className="otp-box">
        <div className="otp-label">Verification code</div>
        <div className="otp-digits">
          {['4','7','2','9'].map((d, i) => <span key={i} className="otp-digit">{d}</span>)}
        </div>
        <div className="otp-hint">Show this to the driver</div>
      </div>
    </div>
  );
}

function WeighingCard({ live, estimate }) {
  const diff = ((live - estimate) / estimate) * 100;
  return (
    <div className="tracker-card stage-card">
      <div className="weigh-icon-row">
        <div className="stage-card-icon weighing-icon"><Icon name="scale" size={20}/></div>
        <div className="weigh-pulse">LIVE</div>
      </div>
      <div className="stage-card-title">Weighing in progress</div>
      <div className="stage-card-body">Reading from certified digital scale</div>

      <div className="weigh-display">
        <div className="weigh-num">{live.toFixed(2)}</div>
        <div className="weigh-unit">kg</div>
      </div>

      <div className="weigh-bar">
        <div className="weigh-bar-fill" style={{width: `${Math.min(100, (live / (estimate * 1.2)) * 100)}%`}}/>
        <div className="weigh-bar-mark" style={{left: `${(estimate / (estimate * 1.2)) * 100}%`}}>
          <span className="mark-lbl">est {estimate}kg</span>
        </div>
      </div>

      <div className="weigh-meta">
        <div>
          <div className="wm-l">Estimated</div>
          <div className="wm-v">{estimate.toFixed(2)} kg</div>
        </div>
        <div>
          <div className="wm-l">Difference</div>
          <div className={`wm-v ${diff >= 0 ? 'pos' : 'neg'}`}>{diff >= 0 ? '+' : ''}{diff.toFixed(1)}%</div>
        </div>
        <div>
          <div className="wm-l">Live payout</div>
          <div className="wm-v">{formatRp(live * 5500)}</div>
        </div>
      </div>
    </div>
  );
}

function QualityCard() {
  const checks = [
    { l: 'Material type', v: 'PET plastic', ok: true },
    { l: 'Cleanliness', v: 'Grade A · clean', ok: true },
    { l: 'Contamination', v: 'None detected', ok: true },
    { l: 'Sorting bonus', v: '+3% applied', ok: true },
  ];
  return (
    <div className="tracker-card stage-card">
      <div className="stage-card-icon quality-icon"><Icon name="shield" size={20}/></div>
      <div className="stage-card-title">Quality verified</div>
      <div className="stage-card-body">All checks passed — your material qualifies for a clean-grade bonus.</div>

      <div className="quality-grid">
        {checks.map((c, i) => (
          <div className="quality-row" key={i}>
            <div className="quality-check"><Icon name="check" size={11} strokeWidth={3} stroke="#fff"/></div>
            <div className="quality-text">
              <div className="ql">{c.l}</div>
              <div className="qv">{c.v}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="proof-photos">
        <div className="proof-label">Proof photos</div>
        <div className="proof-row">
          {['145 165','120 90','100 60'].map((hue, i) => (
            <div key={i} className="proof-thumb" style={{background:`linear-gradient(135deg, oklch(0.75 0.04 ${hue.split(' ')[0]}), oklch(0.55 0.04 ${hue.split(' ')[1]}))`}}/>
          ))}
        </div>
      </div>
    </div>
  );
}

function PayoutCard({ actualWeight, unitPrice, done }) {
  const total = actualWeight * unitPrice;
  return (
    <div className="tracker-card stage-card payout-card">
      <div className="stage-card-icon payout-icon"><Icon name="wallet" size={20}/></div>
      <div className="stage-card-title">{done ? 'Order completed' : 'Payout processed'}</div>
      <div className="stage-card-body">
        {done ? 'Thanks for recycling — your wallet has been credited.' : 'Funds transferred to your EcoCycle wallet.'}
      </div>

      <div className="payout-amount">
        <span className="cur">Rp</span>{Math.round(total).toLocaleString('id-ID')}
      </div>
      <div className="payout-sub">{actualWeight.toFixed(2)} kg × {formatRp(unitPrice)}/kg</div>

      <div className="impact-row">
        <div className="impact-stat">
          <div className="iv">5.9 kg</div>
          <div className="il">CO₂ saved</div>
        </div>
        <div className="impact-stat">
          <div className="iv">+12</div>
          <div className="il">Eco points</div>
        </div>
        <div className="impact-stat">
          <div className="iv">23</div>
          <div className="il">Total orders</div>
        </div>
      </div>

      {done && (
        <div className="rate-row">
          <div style={{fontSize:13, fontWeight:600, marginBottom:8}}>Rate this pickup</div>
          <div style={{display:'flex', gap:6}}>
            {[1,2,3,4,5].map(n => <Icon key={n} name="star" size={22}/>)}
          </div>
        </div>
      )}
    </div>
  );
}

// ───────────────────────────────────────── WALLET
function WalletScreen({ go, balance, t }) {
  const [tab, setTab] = React.useState('all');

  const txs = [
    { id: 1, type: 'in',  t: 'Plastic — 4.2 kg', d: 'Today, 10:14 · Budi Recycling', a: 23100, s: 'Paid' },
    { id: 2, type: 'out', t: 'Withdraw to BCA', d: 'Yesterday, 17:02', a: -150000, s: 'Completed' },
    { id: 3, type: 'in',  t: 'Aluminum — 1.8 kg', d: 'Apr 30 · Sari Daur Ulang', a: 33300, s: 'Paid' },
    { id: 4, type: 'in',  t: 'Cardboard — 8.0 kg', d: 'Apr 28 · Eco Tani', a: 19200, s: 'Paid' },
    { id: 5, type: 'in',  t: 'Referral bonus', d: 'Apr 27 · From Maya', a: 10000, s: 'Bonus' },
    { id: 6, type: 'out', t: 'GoPay top-up', d: 'Apr 25', a: -50000, s: 'Completed' },
  ].filter(x => tab === 'all' || (tab === 'in' && x.type === 'in') || (tab === 'out' && x.type === 'out'));

  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div style={{fontSize:13, fontWeight:700, fontSize:18, letterSpacing:'-0.02em'}}>Wallet</div>
          <div style={{display:'flex', gap:8}}>
            <div className="icon-btn"><Icon name="qr" size={18}/></div>
            <div className="icon-btn"><Icon name="bell" size={18}/></div>
          </div>
        </div>

        <div className="wallet-hero">
          <div className="lbl">Available balance</div>
          <div className="big"><span className="cur">Rp</span>{Math.round(balance).toLocaleString('id-ID')}</div>
          <div className="pending">
            <Icon name="clock" size={13}/>
            Rp 19.250 pending — order #ECC-04827
          </div>
          <div className="wallet-actions">
            <button className="wa-btn"><Icon name="bank" size={18}/>Withdraw</button>
            <button className="wa-btn"><Icon name="send" size={18}/>Transfer</button>
            <button className="wa-btn"><Icon name="qr" size={18}/>Pay</button>
          </div>
        </div>

        <div className="section-head" style={{marginTop:18}}>
          <h3>Transactions</h3>
          <span className="more">Export</span>
        </div>

        <div className="tab-row">
          {[{id:'all',l:'All'}, {id:'in',l:'Earnings'}, {id:'out',l:'Withdrawals'}].map(x => (
            <button key={x.id} className={`tab-btn ${tab === x.id ? 'active' : ''}`} onClick={() => setTab(x.id)}>{x.l}</button>
          ))}
        </div>

        <div className="tx-list">
          {txs.map(x => (
            <div className="tx-row" key={x.id}>
              <div className={`tx-icon ${x.type}`}>
                <Icon name={x.type === 'in' ? 'arrow-down-right' : 'arrow-up-right'} size={18} strokeWidth={2.4}/>
              </div>
              <div className="tx-info">
                <div className="t">{x.t}</div>
                <div className="d">{x.d}</div>
              </div>
              <div className="tx-amt">
                <div className={`v ${x.type}`}>
                  {x.type === 'in' ? '+' : '−'}{formatRp(Math.abs(x.a))}
                </div>
                <div className="s">{x.s}</div>
              </div>
            </div>
          ))}
        </div>

        <div style={{height:8}}/>
      </div>
    </div>
  );
}

// ───────────────────────────────────────── ADMIN
function AdminScreen({ go, t }) {
  const [orders, setOrders] = React.useState([
    { id: 'ECC-04827', user: 'Aria P.', avatar: 'AP', material: 'Plastic', est: 3.5, actual: 3.7, status: 'weighing', address: 'Jl. Kemang Raya 14' },
    { id: 'ECC-04822', user: 'Maya K.', avatar: 'MK', material: 'Aluminum', est: 2.0, actual: null, status: 'pending', address: 'Jl. Cipete 8' },
    { id: 'ECC-04819', user: 'Reza H.', avatar: 'RH', material: 'Copper', est: 1.2, actual: 1.0, status: 'weighing', address: 'Jl. Senopati 22' },
  ]);

  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div>
            <div style={{fontSize:11, color:'var(--ink-3)', fontWeight:600, letterSpacing:'0.04em', textTransform:'uppercase'}}>Collector dashboard</div>
            <div style={{fontSize:18, fontWeight:700, letterSpacing:'-0.02em', marginTop:2}}>Budi Recycling</div>
          </div>
          <div style={{display:'flex',gap:8}}>
            <div className="icon-btn"><Icon name="map" size={18}/></div>
            <div className="icon-btn"><Icon name="chart" size={18}/></div>
          </div>
        </div>

        <div className="admin-stats">
          <div className="admin-stat">
            <div className="l">Today's pickups</div>
            <div className="v">8</div>
            <div className="t">+3 vs yesterday</div>
          </div>
          <div className="admin-stat">
            <div className="l">Revenue</div>
            <div className="v">{formatRp(842500)}</div>
            <div className="t">+12% week</div>
          </div>
          <div className="admin-stat">
            <div className="l">Pending</div>
            <div className="v">{orders.filter(o=>o.status==='pending').length}</div>
            <div className="t" style={{color:'oklch(0.55 0.16 60)'}}>Needs review</div>
          </div>
          <div className="admin-stat">
            <div className="l">Rating</div>
            <div className="v" style={{display:'flex',alignItems:'center',gap:6}}>4.9 <Icon name="star" size={16} stroke="oklch(0.7 0.16 80)"/></div>
            <div className="t">230 reviews</div>
          </div>
        </div>

        <div className="section-head"><h3>Pickup queue</h3><span className="more">Manage</span></div>

        {orders.map(o => {
          const mat = MATERIALS.find(m => m.id === o.material.toLowerCase()) || MATERIALS[0];
          const estPay = mat.unit * o.est;
          const actPay = o.actual ? mat.unit * o.actual : null;
          const diff = o.actual ? ((o.actual - o.est) / o.est * 100).toFixed(1) : null;
          return (
            <div className="admin-card" key={o.id}>
              <div className="ac-head">
                <div className="ac-avatar">{o.avatar}</div>
                <div style={{flex:1, minWidth:0}}>
                  <div className="ac-name">{o.user} · {o.material}</div>
                  <div className="ac-meta">{o.id} · <Icon name="pin" size={10}/> {o.address}</div>
                </div>
                <div className={`ac-status ${o.status}`}>{o.status}</div>
              </div>

              <div className="weight-compare">
                <div className="wc-col">
                  <div className="l">Estimated</div>
                  <div className="v">{o.est} kg</div>
                  <div style={{fontSize:11, color:'var(--ink-3)', marginTop:2}}>{formatRp(estPay)}</div>
                </div>
                <div className="wc-arrow"><Icon name="arrow-right" size={20}/></div>
                <div className="wc-col actual">
                  <div className="l">Actual</div>
                  <div className="v">{o.actual ? `${o.actual} kg` : '—'}</div>
                  <div style={{fontSize:11, color:'var(--accent-deep)', marginTop:2, fontWeight:600}}>
                    {actPay ? formatRp(actPay) : 'Pending'}
                    {diff != null && <span style={{color: diff > 0 ? 'var(--accent-deep)' : 'oklch(0.6 0.18 30)'}}> · {diff > 0 ? '+' : ''}{diff}%</span>}
                  </div>
                </div>
              </div>

              <div className="ac-actions">
                <button className="btn-mini reject"><Icon name="x" size={14} strokeWidth={2.5}/></button>
                <button className="btn-mini negotiate">Negotiate</button>
                <button className="btn-mini accept"><Icon name="check" size={14} strokeWidth={2.5}/> Accept</button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

window.TrackingScreen = TrackingScreen;
window.WalletScreen = WalletScreen;
window.AdminScreen = AdminScreen;
