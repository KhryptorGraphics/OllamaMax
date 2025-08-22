import e,{useState as r,useEffect as n}from"react";import t from"styled-components";var a,o={exports:{}},i={};var l,s={};
/**
 * @license React
 * react-jsx-runtime.development.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */"production"===process.env.NODE_ENV?o.exports=function(){if(a)return i;a=1;var e=Symbol.for("react.transitional.element"),r=Symbol.for("react.fragment");function n(r,n,t){var a=null;if(void 0!==t&&(a=""+t),void 0!==n.key&&(a=""+n.key),"key"in n)for(var o in t={},n)"key"!==o&&(t[o]=n[o]);else t=n;return n=t.ref,{$$typeof:e,type:r,key:a,ref:void 0!==n?n:null,props:t}}return i.Fragment=r,i.jsx=n,i.jsxs=n,i}():o.exports=(l||(l=1,"production"!==process.env.NODE_ENV&&function(){function r(e){if(null==e)return null;if("function"==typeof e)return e.$$typeof===O?null:e.displayName||e.name||null;if("string"==typeof e)return e;switch(e){case b:return"Fragment";case y:return"Profiler";case m:return"StrictMode";case h:return"Suspense";case j:return"SuspenseList";case _:return"Activity"}if("object"==typeof e)switch("number"==typeof e.tag&&console.error("Received an unexpected object in getComponentNameFromType(). This is likely a bug in React. Please file an issue."),e.$$typeof){case f:return"Portal";case x:return(e.displayName||"Context")+".Provider";case g:return(e._context.displayName||"Context")+".Consumer";case v:var n=e.render;return(e=e.displayName)||(e=""!==(e=n.displayName||n.name||"")?"ForwardRef("+e+")":"ForwardRef"),e;case k:return null!==(n=e.displayName||null)?n:r(e.type)||"Memo";case S:n=e._payload,e=e._init;try{return r(e(n))}catch(e){}}return null}function n(e){return""+e}function t(e){try{n(e);var r=!1}catch(e){r=!0}if(r){var t=(r=console).error,a="function"==typeof Symbol&&Symbol.toStringTag&&e[Symbol.toStringTag]||e.constructor.name||"Object";return t.call(r,"The provided key is an unsupported type %s. This value must be coerced to a string before using it here.",a),n(e)}}function a(e){if(e===b)return"<>";if("object"==typeof e&&null!==e&&e.$$typeof===S)return"<...>";try{var n=r(e);return n?"<"+n+">":"<...>"}catch(e){return"<...>"}}function o(){return Error("react-stack-top-frame")}function i(){var e=r(this.type);return R[e]||(R[e]=!0,console.error("Accessing element.ref was removed in React 19. ref is now a regular prop. It will be removed from the JSX Element type in a future release.")),void 0!==(e=this.props.ref)?e:null}function l(e,n,a,o,l,s,p,f){var b,m=n.children;if(void 0!==m)if(o)if($(m)){for(o=0;o<m.length;o++)c(m[o]);Object.freeze&&Object.freeze(m)}else console.error("React.jsx: Static children should always be an array. You are likely explicitly calling React.jsxs or React.jsxDEV. Use the Babel transform instead.");else c(m);if(N.call(n,"key")){m=r(e);var y=Object.keys(n).filter(function(e){return"key"!==e});o=0<y.length?"{key: someKey, "+y.join(": ..., ")+": ...}":"{key: someKey}",C[m+o]||(y=0<y.length?"{"+y.join(": ..., ")+": ...}":"{}",console.error('A props object containing a "key" prop is being spread into JSX:\n  let props = %s;\n  <%s {...props} />\nReact keys must be passed directly to JSX without using spread:\n  let props = %s;\n  <%s key={someKey} {...props} />',o,m,y,m),C[m+o]=!0)}if(m=null,void 0!==a&&(t(a),m=""+a),function(e){if(N.call(e,"key")){var r=Object.getOwnPropertyDescriptor(e,"key").get;if(r&&r.isReactWarning)return!1}return void 0!==e.key}(n)&&(t(n.key),m=""+n.key),"key"in n)for(var g in a={},n)"key"!==g&&(a[g]=n[g]);else a=n;return m&&function(e,r){function n(){d||(d=!0,console.error("%s: `key` is not a prop. Trying to access it will result in `undefined` being returned. If you need to access the same value within the child component, you should pass it as a different prop. (https://react.dev/link/special-props)",r))}n.isReactWarning=!0,Object.defineProperty(e,"key",{get:n,configurable:!0})}(a,"function"==typeof e?e.displayName||e.name||"Unknown":e),function(e,r,n,t,a,o,l,s){return n=o.ref,e={$$typeof:u,type:e,key:r,props:o,_owner:a},null!==(void 0!==n?n:null)?Object.defineProperty(e,"ref",{enumerable:!1,get:i}):Object.defineProperty(e,"ref",{enumerable:!1,value:null}),e._store={},Object.defineProperty(e._store,"validated",{configurable:!1,enumerable:!1,writable:!0,value:0}),Object.defineProperty(e,"_debugInfo",{configurable:!1,enumerable:!1,writable:!0,value:null}),Object.defineProperty(e,"_debugStack",{configurable:!1,enumerable:!1,writable:!0,value:l}),Object.defineProperty(e,"_debugTask",{configurable:!1,enumerable:!1,writable:!0,value:s}),Object.freeze&&(Object.freeze(e.props),Object.freeze(e)),e}(e,m,s,0,null===(b=w.A)?null:b.getOwner(),a,p,f)}function c(e){"object"==typeof e&&null!==e&&e.$$typeof===u&&e._store&&(e._store.validated=1)}var d,p=e,u=Symbol.for("react.transitional.element"),f=Symbol.for("react.portal"),b=Symbol.for("react.fragment"),m=Symbol.for("react.strict_mode"),y=Symbol.for("react.profiler"),g=Symbol.for("react.consumer"),x=Symbol.for("react.context"),v=Symbol.for("react.forward_ref"),h=Symbol.for("react.suspense"),j=Symbol.for("react.suspense_list"),k=Symbol.for("react.memo"),S=Symbol.for("react.lazy"),_=Symbol.for("react.activity"),O=Symbol.for("react.client.reference"),w=p.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,N=Object.prototype.hasOwnProperty,$=Array.isArray,E=console.createTask?console.createTask:function(){return null},R={},T=(p={react_stack_bottom_frame:function(e){return e()}}).react_stack_bottom_frame.bind(p,o)(),P=E(a(o)),C={};s.Fragment=b,s.jsx=function(e,r,n,t,o){var i=1e4>w.recentlyCreatedOwnerStacks++;return l(e,r,n,!1,0,o,i?Error("react-stack-top-frame"):T,i?E(a(e)):P)},s.jsxs=function(e,r,n,t,o){var i=1e4>w.recentlyCreatedOwnerStacks++;return l(e,r,n,!0,0,o,i?Error("react-stack-top-frame"):T,i?E(a(e)):P)}}()),s);var c=o.exports;const d=t.button`
  --bg: var(--omx-color-brand-500);
  --bg-hover: var(--omx-color-brand-600);
  --fg: #fff;
  --border: transparent;

  ${e=>"secondary"===e.$variant&&"--bg: var(--omx-color-bg-surface); --fg: var(--omx-color-text-default); --border: var(--omx-color-text-muted);"}
  ${e=>"danger"===e.$variant&&"--bg: #ef4444; --bg-hover: #dc2626;"}
  ${e=>"ghost"===e.$variant&&"--bg: transparent; --bg-hover: rgba(0,0,0,0.05); --fg: var(--omx-color-text-default);"}

  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--omx-spacing-2) var(--omx-spacing-4);
  border-radius: var(--omx-radius-md);
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  transition: background 150ms ease;

  &:hover { background: var(--bg-hover); }
  &:focus-visible { outline: 2px solid var(--omx-color-brand-500); outline-offset: 2px; }
`,p=({variant:e="primary",children:r,...n})=>c.jsx(d,{className:"omx-v2",$variant:e,...n,children:r}),u=t.header`
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--omx-spacing-3) var(--omx-spacing-4);
  background: var(--omx-color-bg-default, #fff);
  border-bottom: 1px solid rgba(0,0,0,0.06);
`,f=t.div`
  display: flex; align-items: center; gap: var(--omx-spacing-2);
  font-weight: 600; color: var(--omx-color-text-default, #0f172a);
`,b=t.nav`
  display: none; gap: var(--omx-spacing-4);
  @media (min-width: 768px) { display: flex; }
`,m=t.a`
  color: #2563eb; text-decoration: none;
  &:hover { text-decoration: underline; }
`,y=t.div`
  display: flex; align-items: center; gap: var(--omx-spacing-3);
`,g=t.button`
  background: transparent; border: none; cursor: pointer; font-size: 1.25rem;
`,x=({brand:e="OllamaMax",links:r=[],onToggleMenu:n,user:t})=>c.jsxs(u,{className:"omx-v2",role:"banner",children:[c.jsxs(f,{children:[c.jsx(g,{"aria-label":"Open menu",onClick:n,style:{display:"inline-flex"},children:c.jsx("span",{"aria-hidden":!0,children:"☰"})}),c.jsx("span",{children:e})]}),c.jsx(b,{"aria-label":"Primary",children:r.map(e=>c.jsx(m,{href:e.href,children:e.label},e.href))}),c.jsxs(y,{children:[c.jsx(g,{"aria-label":"Notifications",children:c.jsx("span",{"aria-hidden":!0,children:"🔔"})}),t?.name&&c.jsxs("div",{"aria-label":"User menu",children:[t.name," ",t.onLogout&&c.jsx("button",{onClick:t.onLogout,"aria-label":"Logout",style:{marginLeft:8},children:"Logout"})]})]})]}),v=t.aside`
  position: sticky; top: 0; height: 100dvh; z-index: 40;
  width: ${e=>e.collapsed?"64px":"240px"};
  background: var(--omx-color-bg-subtle, #0ea5e91a);
  border-right: 1px solid rgba(0,0,0,0.06);
  transition: width .2s ease;
  display: flex; flex-direction: column; gap: 4px; padding: 8px;
`,h=t.a`
  display: flex; align-items: center; gap: 8px; padding: 8px; border-radius: 8px;
  color: ${e=>e.active?"#0f172a":"#334155"}; text-decoration: none;
  background: ${e=>e.active?"rgba(14,165,233,0.12)":"transparent"};
  &:hover { background: rgba(14,165,233,0.10); }
  > span.label { display: ${e=>e.collapsed?"none":"inline"}; }
`,j=t.button`
  background: transparent; border: none; cursor: pointer; padding: 8px; text-align: left;
`,k=({items:e,activeHref:t,collapsedKey:a="omx-sidenav",defaultCollapsed:o=!1,onNavigate:i})=>{const[l,s]=r(o);n(()=>{const e=localStorage.getItem(a);null!=e&&s("1"===e)},[a]);return c.jsxs(v,{className:"omx-v2",role:"navigation","aria-label":"Sidebar",collapsed:l,children:[c.jsx(j,{onClick:()=>{s(e=>{const r=!e;try{localStorage.setItem(a,r?"1":"0")}catch{}return r})},"aria-label":l?"Expand sidebar":"Collapse sidebar",children:l?"»":"«"}),c.jsx("div",{role:"list",children:e.map(e=>c.jsxs(h,{href:e.href,active:t===e.href,collapsed:l,onClick:r=>{i&&(r.preventDefault(),i(e.href))},role:"listitem",children:[e.icon&&c.jsx("span",{"aria-hidden":!0,children:e.icon}),c.jsx("span",{className:"label",children:e.label})]},e.href))})]})},S=t.nav`
  font-size: 0.875rem; color: #475569; padding: 8px 16px;
`,_=t.ol`
  display: flex; align-items: center; gap: 8px; list-style: none; padding: 0; margin: 0;
`,O=t.a`
  color: #2563eb; text-decoration: none; &:hover{ text-decoration: underline; }
`,w=({items:e,"aria-label":r="Breadcrumb"})=>c.jsx(S,{"aria-label":r,className:"omx-v2",children:c.jsx(_,{children:e.map((r,n)=>{const t=n===e.length-1;return c.jsxs("li",{"aria-current":t?"page":void 0,children:[r.href&&!t?c.jsx(O,{href:r.href,children:r.label}):c.jsx("span",{children:r.label}),!t&&c.jsx("span",{"aria-hidden":!0,style:{margin:"0 4px"},children:"/"})]},n)})})});export{w as Breadcrumbs,p as Button,x as Header,k as SideNav};
//# sourceMappingURL=index.esm.js.map
