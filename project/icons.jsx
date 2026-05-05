/* global React */
// Icon set — minimal line icons with consistent stroke

const Icon = ({ name, size = 22, stroke = 'currentColor', fill = 'none', strokeWidth = 2 }) => {
  const props = {
    width: size, height: size, viewBox: '0 0 24 24',
    fill, stroke, strokeWidth, strokeLinecap: 'round', strokeLinejoin: 'round'
  };
  switch (name) {
    case 'home': return (<svg {...props}><path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V20a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1V9.5"/></svg>);
    case 'home-fill': return (<svg {...props} fill="currentColor" stroke="none"><path d="M3 10.5 12 3l9 7.5V20a1 1 0 0 1-1 1h-5v-6h-4v6H6a1 1 0 0 1-1-1V10.5z"/></svg>);
    case 'map': return (<svg {...props}><path d="M9 4 3 6v14l6-2 6 2 6-2V4l-6 2-6-2z"/><path d="M9 4v14M15 6v14"/></svg>);
    case 'map-fill': return (<svg {...props} fill="currentColor" stroke="none"><path d="M9 4 3 6v14l6-2v-14zM15 6l6-2v14l-6 2V6zM9 4v14l6 2V6L9 4z" opacity=".5"/><path d="M9 4 3 6v14l6-2 6 2 6-2V4l-6 2-6-2z" fill="none" stroke="currentColor" strokeWidth="2"/><path d="M9 4v14M15 6v14" stroke="currentColor" strokeWidth="2"/></svg>);
    case 'wallet': return (<svg {...props}><path d="M3 7a2 2 0 0 1 2-2h12l4 4v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/><path d="M16 13h2"/></svg>);
    case 'wallet-fill': return (<svg {...props} fill="currentColor" stroke="none"><path d="M3 7a2 2 0 0 1 2-2h12l4 4v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/><circle cx="17" cy="13" r="1.4" fill="#fff"/></svg>);
    case 'user': return (<svg {...props}><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4.4 3.6-8 8-8s8 3.6 8 8"/></svg>);
    case 'user-fill': return (<svg {...props} fill="currentColor" stroke="none"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4.4 3.6-8 8-8s8 3.6 8 8"/></svg>);
    case 'plus': return (<svg {...props}><path d="M12 5v14M5 12h14"/></svg>);
    case 'bell': return (<svg {...props}><path d="M6 8a6 6 0 1 1 12 0c0 5 2 6 2 8H4c0-2 2-3 2-8z"/><path d="M10 21h4"/></svg>);
    case 'arrow-right': return (<svg {...props}><path d="M5 12h14M13 6l6 6-6 6"/></svg>);
    case 'arrow-left': return (<svg {...props}><path d="M19 12H5M11 6l-6 6 6 6"/></svg>);
    case 'chevron-right': return (<svg {...props}><path d="M9 6l6 6-6 6"/></svg>);
    case 'chevron-down': return (<svg {...props}><path d="M6 9l6 6 6-6"/></svg>);
    case 'leaf': return (<svg {...props}><path d="M11 20A7 7 0 0 1 4 13V5l8 3a7 7 0 0 1 7 7v5h-8z"/><path d="M4 5l11 11"/></svg>);
    case 'leaf-fill': return (<svg {...props} fill="currentColor" stroke="none"><path d="M20 4c0 9-5 14-12 14-2 0-4-1-4-1s4-1 7-5c-2 1-5 1-5 1s2-3 6-5c-3 0-5 1-5 1s4-5 13-5z"/></svg>);
    case 'truck': return (<svg {...props}><path d="M3 7h11v9H3zM14 11h4l3 3v2h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/></svg>);
    case 'pin': return (<svg {...props}><path d="M12 22s7-7.5 7-13a7 7 0 1 0-14 0c0 5.5 7 13 7 13z"/><circle cx="12" cy="9" r="2.5"/></svg>);
    case 'star': return (<svg {...props} fill="currentColor"><path d="m12 3 2.7 5.5 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.8 1-6.1L3.2 9.4l6.1-.9L12 3z"/></svg>);
    case 'search': return (<svg {...props}><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>);
    case 'filter': return (<svg {...props}><path d="M3 5h18M6 12h12M10 19h4"/></svg>);
    case 'phone': return (<svg {...props}><path d="M5 4h3l2 5-2.5 1.5a11 11 0 0 0 6 6L15 14l5 2v3a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"/></svg>);
    case 'message': return (<svg {...props}><path d="M4 5h16v12H8l-4 4V5z"/></svg>);
    case 'check': return (<svg {...props}><path d="m5 12 5 5L20 7"/></svg>);
    case 'check-circle': return (<svg {...props} fill="currentColor" stroke="none"><circle cx="12" cy="12" r="10"/><path d="m8 12 3 3 5-6" stroke="#fff" strokeWidth="2.5" fill="none" strokeLinecap="round" strokeLinejoin="round"/></svg>);
    case 'x': return (<svg {...props}><path d="M6 6l12 12M6 18 18 6"/></svg>);
    case 'camera': return (<svg {...props}><path d="M4 8h3l2-2h6l2 2h3v11H4z"/><circle cx="12" cy="13" r="3.5"/></svg>);
    case 'image': return (<svg {...props}><rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="9" cy="10" r="1.5"/><path d="m3 17 5-5 4 4 3-3 6 6"/></svg>);
    case 'home-pickup': return (<svg {...props}><path d="M3 11 12 4l9 7"/><path d="M5 10v9h14v-9"/><path d="M10 19v-5h4v5"/></svg>);
    case 'box': return (<svg {...props}><path d="m3 7 9-4 9 4v10l-9 4-9-4V7z"/><path d="m3 7 9 4 9-4M12 11v10"/></svg>);
    case 'qr': return (<svg {...props}><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><path d="M14 14h3v3M21 14v3M14 21h3M21 17v4"/></svg>);
    case 'arrow-down-right': return (<svg {...props}><path d="M7 7l10 10M17 9v8h-8"/></svg>);
    case 'arrow-up-right': return (<svg {...props}><path d="M7 17 17 7M9 7h8v8"/></svg>);
    case 'send': return (<svg {...props}><path d="m3 11 18-8-7 18-3-7-8-3z"/></svg>);
    case 'bank': return (<svg {...props}><path d="M3 10 12 4l9 6"/><path d="M5 10v8M19 10v8M9 10v8M15 10v8M3 20h18"/></svg>);
    case 'sparkles': return (<svg {...props}><path d="M12 4v6M12 14v6M4 12h6M14 12h6"/><path d="m6 6 1.5 1.5M16.5 16.5 18 18M6 18l1.5-1.5M16.5 7.5 18 6"/></svg>);
    case 'recycle': return (<svg {...props}><path d="M7 6 5 9h4l-2-3zm10 0-2 3h4l-2-3zM12 18l-2-3h4l-2 3z"/><path d="M9 9 6 14l3 1M15 9l3 5-3 1M14 15l-2 3-2-3"/></svg>);
    case 'eye': return (<svg {...props}><path d="M2 12s4-8 10-8 10 8 10 8-4 8-10 8-10-8-10-8z"/><circle cx="12" cy="12" r="3"/></svg>);
    case 'shield': return (<svg {...props}><path d="M12 3 4 6v6c0 5 3.5 8.5 8 9 4.5-.5 8-4 8-9V6l-8-3z"/><path d="m9 12 2 2 4-4"/></svg>);
    case 'clock': return (<svg {...props}><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>);
    case 'trend-up': return (<svg {...props}><path d="m3 17 6-6 4 4 8-8"/><path d="M14 7h7v7"/></svg>);
    case 'navigation': return (<svg {...props}><path d="m3 11 18-8-8 18-2-8-8-2z"/></svg>);
    case 'package': return (<svg {...props}><path d="m3 7 9-4 9 4v10l-9 4-9-4V7z"/><path d="M3 7l9 4 9-4M12 11v10M7.5 5l9 4"/></svg>);
    case 'scale': return (<svg {...props}><path d="M12 3v18M5 7l-2 6h6l-2-6zM19 7l-2 6h6l-2-6z"/><path d="M5 21h14"/></svg>);
    case 'chart': return (<svg {...props}><path d="M4 19V5M4 19h16M8 15v-4M12 15V8M16 15v-6"/></svg>);
    default: return null;
  }
};

window.Icon = Icon;

// Material catalog — unified monochrome icon + price/kg
const MaterialIcon = ({ name, size = 22 }) => {
  const p = { width: size, height: size, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 1.6, strokeLinecap: 'round', strokeLinejoin: 'round' };
  switch (name) {
    case 'plastic':   return (<svg {...p}><path d="M8 4h8l-1 4h-6zM7 8h10l1 11a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2z"/><path d="M10 13v4M14 13v4"/></svg>);
    case 'cardboard': return (<svg {...p}><path d="M3 7l9-4 9 4v10l-9 4-9-4V7z"/><path d="M3 7l9 4 9-4M12 11v10"/></svg>);
    case 'paper':     return (<svg {...p}><path d="M6 3h9l3 3v15H6z"/><path d="M15 3v3h3M9 11h6M9 15h6M9 19h4"/></svg>);
    case 'aluminum':  return (<svg {...p}><path d="M7 4h10l-1 16H8z"/><path d="M7 8h10M9 4v16M15 4v16"/></svg>);
    case 'copper':    return (<svg {...p}><circle cx="12" cy="12" r="8"/><circle cx="12" cy="12" r="4"/><path d="M12 4v2M12 18v2M4 12h2M18 12h2"/></svg>);
    case 'steel':     return (<svg {...p}><path d="M4 6h16v4l-8 10L4 10z"/><path d="M4 10h16M9 6v4M15 6v4"/></svg>);
    case 'glass':     return (<svg {...p}><path d="M7 3h10l-1 9a4 4 0 0 1-8 0z"/><path d="M12 16v5M9 21h6"/></svg>);
    case 'ewaste':    return (<svg {...p}><rect x="5" y="7" width="14" height="11" rx="1.5"/><path d="M9 7V4M15 7V4M9 18v3M15 18v3"/><circle cx="12" cy="12" r="2"/></svg>);
    default: return null;
  }
};
window.MaterialIcon = MaterialIcon;

const MATERIALS = [
  { id: 'plastic',   name: 'Plastic',   unit: 5500 },
  { id: 'cardboard', name: 'Cardboard', unit: 2400 },
  { id: 'paper',     name: 'Paper',     unit: 3200 },
  { id: 'aluminum',  name: 'Aluminum',  unit: 18500 },
  { id: 'copper',    name: 'Copper',    unit: 92000 },
  { id: 'steel',     name: 'Steel',     unit: 6800 },
  { id: 'glass',     name: 'Glass',     unit: 1200 },
  { id: 'ewaste',    name: 'E-waste',   unit: 22000 },
];

window.MATERIALS = MATERIALS;

const formatRp = (n) => 'Rp ' + Math.round(n).toLocaleString('id-ID');
window.formatRp = formatRp;

Object.assign(window, { Icon, MATERIALS, formatRp });
