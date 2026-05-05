/* global React, ReactDOM, Icon, HomeScreen, SellScreen, MapScreen, TrackingScreen, WalletScreen, AdminScreen,
   IOSDevice, useTweaks, TweaksPanel, TweakSection, TweakRadio, TweakSlider, TweakToggle, TweakColor */

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accent": "#22c55e",
  "background": "mesh",
  "blur": 20,
  "radius": 18,
  "dark": false
}/*EDITMODE-END*/;

function BottomNav({ active, go }) {
  const items = [
    { id: 'home', label: 'Home', icon: 'home' },
    { id: 'map', label: 'Map', icon: 'map' },
    { id: 'sell', label: 'Sell', fab: true, icon: 'plus' },
    { id: 'wallet', label: 'Wallet', icon: 'wallet' },
    { id: 'profile', label: 'Profile', icon: 'user' },
  ];
  return (
    <div className="bottom-nav">
      {items.map(it => {
        if (it.fab) {
          return (
            <div key={it.id} className="nav-item fab" onClick={() => go(it.id)}>
              <div className="fab-circle"><Icon name="plus" size={22} strokeWidth={2.5} stroke="#fff"/></div>
            </div>
          );
        }
        const isActive = active === it.id;
        return (
          <div key={it.id} className={`nav-item ${isActive ? 'active' : ''}`} onClick={() => go(it.id)}>
            <div className="pill"/>
            <Icon name={isActive ? `${it.icon}-fill` : it.icon} size={22}/>
            <div className="label">{it.label}</div>
          </div>
        );
      })}
    </div>
  );
}

function App() {
  const [t, setTweak] = useTweaks(TWEAK_DEFAULTS);
  const [screen, setScreen] = React.useState('home');
  const [balance, setBalance] = React.useState(347500);
  const [toast, setToast] = React.useState(null);

  // Tabs that have bottom-nav vs full custom
  const navScreens = ['home', 'map', 'wallet', 'profile'];
  const showNav = navScreens.includes(screen) || screen === 'sell' || screen === 'tracking' || screen === 'admin';

  const go = (s) => {
    setScreen(s);
  };

  const onSubmit = (data) => {
    setBalance(b => b + data.estimate);
    setToast(`Order created · ${data.weight} kg ${data.material}`);
    setTimeout(() => setToast(null), 2200);
    setTimeout(() => setScreen('tracking'), 600);
  };

  // Apply tweaks via CSS vars
  React.useEffect(() => {
    const root = document.documentElement;
    root.style.setProperty('--accent', t.accent);
    // Derive a deeper accent for hover/text
    root.style.setProperty('--glass-blur', `${t.blur}px`);
    root.style.setProperty('--radius', `${t.radius}px`);
    root.style.setProperty('--radius-sm', `${Math.max(8, t.radius - 6)}px`);
    root.style.setProperty('--radius-lg', `${t.radius + 6}px`);
    root.dataset.theme = t.dark ? 'dark' : 'light';
  }, [t]);

  let body = null;
  switch (screen) {
    case 'home': body = <HomeScreen go={go} balance={balance} t={t}/>; break;
    case 'sell': body = <SellScreen go={go} t={t} onSubmit={onSubmit}/>; break;
    case 'map': body = <MapScreen go={go} t={t}/>; break;
    case 'tracking': body = <TrackingScreen go={go} t={t}/>; break;
    case 'wallet': body = <WalletScreen go={go} balance={balance} t={t}/>; break;
    case 'admin': body = <AdminScreen go={go} t={t}/>; break;
    case 'profile':
    case 'notif':
    case 'prices':
      body = <PlaceholderScreen go={go} screen={screen} t={t}/>;
      break;
    default: body = <HomeScreen go={go} balance={balance} t={t}/>;
  }

  // Map active for nav highlighting
  const navActive = ({
    home: 'home', map: 'map', wallet: 'wallet', profile: 'profile',
    sell: 'sell', tracking: 'home', admin: 'home',
  })[screen] || 'home';

  // Determine status bar style — map screen has light backdrop, wallet has green hero
  const statusDark = t.dark;

  return (
    <div className="stage">
      <div className={`stage-bg ${t.background}`}/>

      <div className="device-stage">
        <IOSDevice width={390} height={844} dark={t.dark}>
          <div data-screen-label={screenLabel(screen)} style={{position:'absolute', inset:0}}>
            {body}
            <BottomNav active={navActive} go={go}/>
            {toast && (
              <div className="toast">
                <Icon name="check-circle" size={16} stroke="none" fill="#fff"/>
                {toast}
              </div>
            )}
          </div>
        </IOSDevice>
      </div>

      <TweaksPanel title="Tweaks">
        <TweakSection label="Theme">
          <TweakToggle label="Dark mode" value={t.dark} onChange={v => setTweak('dark', v)}/>
          <TweakColor label="Accent" value={t.accent} onChange={v => setTweak('accent', v)}/>
        </TweakSection>
        <TweakSection label="Glass">
          <TweakSlider label="Blur" value={t.blur} min={0} max={40} unit="px" onChange={v => setTweak('blur', v)}/>
          <TweakSlider label="Radius" value={t.radius} min={8} max={28} unit="px" onChange={v => setTweak('radius', v)}/>
        </TweakSection>
        <TweakSection label="Backdrop">
          <TweakRadio label="Style" value={t.background}
            options={[{value:'mesh',label:'Mesh'},{value:'blobs',label:'Blobs'},{value:'minimal',label:'Minimal'}]}
            onChange={v => setTweak('background', v)}/>
        </TweakSection>
      </TweaksPanel>
    </div>
  );
}

function screenLabel(s) {
  return ({
    home: '01 Home', sell: '02 Sell Scrap', map: '03 Map',
    tracking: '04 Order Tracking', wallet: '05 Wallet',
    admin: '06 Collector Dashboard',
    profile: '07 Profile', notif: 'Notifications', prices: 'All Prices',
  })[s] || s;
}

function ViewSwitcher({ screen, go }) {
  const items = [
    { id: 'home', l: 'Home' },
    { id: 'sell', l: 'Sell' },
    { id: 'map', l: 'Map' },
    { id: 'tracking', l: 'Track' },
    { id: 'wallet', l: 'Wallet' },
    { id: 'admin', l: 'Admin' },
  ];
  return (
    <div style={{
      position:'fixed', top:20, left:'50%', transform:'translateX(-50%)',
      display:'flex', gap:4, padding:4, borderRadius:999,
      background:'rgba(255,255,255,0.7)',
      backdropFilter:'blur(20px) saturate(180%)',
      WebkitBackdropFilter:'blur(20px) saturate(180%)',
      border:'1px solid rgba(255,255,255,0.7)',
      boxShadow:'0 8px 30px rgba(20,60,40,0.12)',
      zIndex:50, fontFamily:'var(--font)',
    }}>
      {items.map(it => {
        const active = screen === it.id || (it.id === 'home' && screen === 'profile');
        return (
          <button key={it.id} onClick={() => go(it.id)}
            style={{
              padding:'6px 14px', borderRadius:999, border:0, cursor:'pointer',
              fontSize:12, fontWeight:600,
              background: active ? 'var(--accent)' : 'transparent',
              color: active ? '#fff' : 'var(--ink-2)',
              fontFamily:'var(--font)',
              transition:'all .15s ease',
            }}>{it.l}</button>
        );
      })}
    </div>
  );
}

function PlaceholderScreen({ go, screen }) {
  const titles = {
    profile: { title: 'Profile', sub: 'Your account, eco-impact, settings' },
    notif: { title: 'Notifications', sub: '3 updates today' },
    prices: { title: 'All scrap prices', sub: 'Live market rates per kg' },
  };
  const m = titles[screen] || { title: screen, sub: '' };

  if (screen === 'profile') return <ProfileScreen go={go}/>;

  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div className="icon-btn" onClick={() => go('home')}><Icon name="arrow-left" size={20}/></div>
          <div style={{fontSize:13, fontWeight:600}}>{m.title}</div>
          <div style={{width:42}}/>
        </div>
        <div className="page-title">{m.title}</div>
        <div className="page-sub">{m.sub}</div>
        <div className="glass" style={{padding:24, textAlign:'center', color:'var(--ink-3)', fontSize:13}}>
          <Icon name="leaf" size={32} stroke="var(--accent)"/>
          <div style={{marginTop:12}}>Screen content here.</div>
        </div>
      </div>
    </div>
  );
}

function ProfileScreen({ go }) {
  return (
    <div className="screen page-enter">
      <div className="screen-bg"/>
      <div className="scroll">
        <div className="top-header">
          <div style={{fontSize:18, fontWeight:700, letterSpacing:'-0.02em'}}>Profile</div>
          <div className="icon-btn"><Icon name="bell" size={18}/></div>
        </div>

        <div className="glass" style={{padding:18, display:'flex', alignItems:'center', gap:14, marginTop:8}}>
          <div className="avatar" style={{width:56, height:56, fontSize:18}}>AR</div>
          <div style={{flex:1}}>
            <div style={{fontSize:16, fontWeight:700}}>Aria Putri</div>
            <div style={{fontSize:12, color:'var(--ink-3)', marginTop:2}}>Eco-warrior · since 2024</div>
          </div>
          <div style={{padding:'4px 10px', borderRadius:999, background:'var(--accent-soft)', color:'var(--accent-deep)', fontSize:11, fontWeight:700, display:'flex', alignItems:'center', gap:4}}>
            <Icon name="leaf" size={11}/>Gold
          </div>
        </div>

        <div className="section-head" style={{marginTop:18}}><h3>Your impact</h3></div>
        <div className="admin-stats">
          <div className="admin-stat">
            <div className="l">Total recycled</div>
            <div className="v">42.6 kg</div>
            <div className="t">23 orders</div>
          </div>
          <div className="admin-stat">
            <div className="l">CO₂ saved</div>
            <div className="v">68 kg</div>
            <div className="t">≈ 4 trees</div>
          </div>
          <div className="admin-stat">
            <div className="l">Earnings</div>
            <div className="v">{window.formatRp(847500)}</div>
            <div className="t">All-time</div>
          </div>
          <div className="admin-stat">
            <div className="l">Rank</div>
            <div className="v">#142</div>
            <div className="t">Jakarta</div>
          </div>
        </div>

        <div className="section-head"><h3>Settings</h3></div>
        {[
          {ico:'user', l:'Personal info'},
          {ico:'pin', l:'Saved addresses'},
          {ico:'bank', l:'Payout methods'},
          {ico:'shield', l:'Privacy & security'},
          {ico:'recycle', l:'Switch to Collector mode', accent:true},
        ].map((r,i) => (
          <div key={i} className="glass" style={{padding:'14px 16px', display:'flex', alignItems:'center', gap:12, marginBottom:8}}>
            <div style={{width:34, height:34, borderRadius:10, background: r.accent ? 'var(--accent-soft)' : 'rgba(255,255,255,0.5)', color: r.accent ? 'var(--accent-deep)' : 'var(--ink-2)', display:'grid', placeItems:'center'}}>
              <Icon name={r.ico} size={16}/>
            </div>
            <div style={{flex:1, fontSize:13, fontWeight: r.accent ? 700 : 500, color: r.accent ? 'var(--accent-deep)' : 'var(--ink)'}}>{r.l}</div>
            <Icon name="chevron-right" size={16} stroke="var(--ink-4)"/>
          </div>
        ))}
      </div>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App/>);
