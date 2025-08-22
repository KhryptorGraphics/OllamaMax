!function(e,r){"object"==typeof exports&&"undefined"!=typeof module?r(exports,require("react"),require("styled-components")):"function"==typeof define&&define.amd?define(["exports","react","styled-components"],r):r((e="undefined"!=typeof globalThis?globalThis:e||self).OmxUI={},e.React,e.styled)}(this,function(e,r,t){"use strict";var n,a={exports:{}},o={};var i,l={};
/**
   * @license React
   * react-jsx-runtime.development.js
   *
   * Copyright (c) Meta Platforms, Inc. and affiliates.
   *
   * This source code is licensed under the MIT license found in the
   * LICENSE file in the root directory of this source tree.
   */"production"===process.env.NODE_ENV?a.exports=function(){if(n)return o;n=1;var e=Symbol.for("react.transitional.element"),r=Symbol.for("react.fragment");function t(r,t,n){var a=null;if(void 0!==n&&(a=""+n),void 0!==t.key&&(a=""+t.key),"key"in t)for(var o in n={},t)"key"!==o&&(n[o]=t[o]);else n=t;return t=n.ref,{$$typeof:e,type:r,key:a,ref:void 0!==t?t:null,props:n}}return o.Fragment=r,o.jsx=t,o.jsxs=t,o}():a.exports=(i||(i=1,"production"!==process.env.NODE_ENV&&function(){function e(r){if(null==r)return null;if("function"==typeof r)return r.$$typeof===O?null:r.displayName||r.name||null;if("string"==typeof r)return r;switch(r){case b:return"Fragment";case y:return"Profiler";case m:return"StrictMode";case h:return"Suspense";case j:return"SuspenseList";case _:return"Activity"}if("object"==typeof r)switch("number"==typeof r.tag&&console.error("Received an unexpected object in getComponentNameFromType(). This is likely a bug in React. Please file an issue."),r.$$typeof){case f:return"Portal";case x:return(r.displayName||"Context")+".Provider";case g:return(r._context.displayName||"Context")+".Consumer";case v:var t=r.render;return(r=r.displayName)||(r=""!==(r=t.displayName||t.name||"")?"ForwardRef("+r+")":"ForwardRef"),r;case k:return null!==(t=r.displayName||null)?t:e(r.type)||"Memo";case S:t=r._payload,r=r._init;try{return e(r(t))}catch(e){}}return null}function t(e){return""+e}function n(e){try{t(e);var r=!1}catch(e){r=!0}if(r){var n=(r=console).error,a="function"==typeof Symbol&&Symbol.toStringTag&&e[Symbol.toStringTag]||e.constructor.name||"Object";return n.call(r,"The provided key is an unsupported type %s. This value must be coerced to a string before using it here.",a),t(e)}}function a(r){if(r===b)return"<>";if("object"==typeof r&&null!==r&&r.$$typeof===S)return"<...>";try{var t=e(r);return t?"<"+t+">":"<...>"}catch(e){return"<...>"}}function o(){return Error("react-stack-top-frame")}function i(){var r=e(this.type);return E[r]||(E[r]=!0,console.error("Accessing element.ref was removed in React 19. ref is now a regular prop. It will be removed from the JSX Element type in a future release.")),void 0!==(r=this.props.ref)?r:null}function s(r,t,a,o,l,s,u,f){var b,m=t.children;if(void 0!==m)if(o)if($(m)){for(o=0;o<m.length;o++)c(m[o]);Object.freeze&&Object.freeze(m)}else console.error("React.jsx: Static children should always be an array. You are likely explicitly calling React.jsxs or React.jsxDEV. Use the Babel transform instead.");else c(m);if(N.call(t,"key")){m=e(r);var y=Object.keys(t).filter(function(e){return"key"!==e});o=0<y.length?"{key: someKey, "+y.join(": ..., ")+": ...}":"{key: someKey}",C[m+o]||(y=0<y.length?"{"+y.join(": ..., ")+": ...}":"{}",console.error('A props object containing a "key" prop is being spread into JSX:\n  let props = %s;\n  <%s {...props} />\nReact keys must be passed directly to JSX without using spread:\n  let props = %s;\n  <%s key={someKey} {...props} />',o,m,y,m),C[m+o]=!0)}if(m=null,void 0!==a&&(n(a),m=""+a),function(e){if(N.call(e,"key")){var r=Object.getOwnPropertyDescriptor(e,"key").get;if(r&&r.isReactWarning)return!1}return void 0!==e.key}(t)&&(n(t.key),m=""+t.key),"key"in t)for(var g in a={},t)"key"!==g&&(a[g]=t[g]);else a=t;return m&&function(e,r){function t(){d||(d=!0,console.error("%s: `key` is not a prop. Trying to access it will result in `undefined` being returned. If you need to access the same value within the child component, you should pass it as a different prop. (https://react.dev/link/special-props)",r))}t.isReactWarning=!0,Object.defineProperty(e,"key",{get:t,configurable:!0})}(a,"function"==typeof r?r.displayName||r.name||"Unknown":r),function(e,r,t,n,a,o,l,s){return t=o.ref,e={$$typeof:p,type:e,key:r,props:o,_owner:a},null!==(void 0!==t?t:null)?Object.defineProperty(e,"ref",{enumerable:!1,get:i}):Object.defineProperty(e,"ref",{enumerable:!1,value:null}),e._store={},Object.defineProperty(e._store,"validated",{configurable:!1,enumerable:!1,writable:!0,value:0}),Object.defineProperty(e,"_debugInfo",{configurable:!1,enumerable:!1,writable:!0,value:null}),Object.defineProperty(e,"_debugStack",{configurable:!1,enumerable:!1,writable:!0,value:l}),Object.defineProperty(e,"_debugTask",{configurable:!1,enumerable:!1,writable:!0,value:s}),Object.freeze&&(Object.freeze(e.props),Object.freeze(e)),e}(r,m,s,0,null===(b=w.A)?null:b.getOwner(),a,u,f)}function c(e){"object"==typeof e&&null!==e&&e.$$typeof===p&&e._store&&(e._store.validated=1)}var d,u=r,p=Symbol.for("react.transitional.element"),f=Symbol.for("react.portal"),b=Symbol.for("react.fragment"),m=Symbol.for("react.strict_mode"),y=Symbol.for("react.profiler"),g=Symbol.for("react.consumer"),x=Symbol.for("react.context"),v=Symbol.for("react.forward_ref"),h=Symbol.for("react.suspense"),j=Symbol.for("react.suspense_list"),k=Symbol.for("react.memo"),S=Symbol.for("react.lazy"),_=Symbol.for("react.activity"),O=Symbol.for("react.client.reference"),w=u.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,N=Object.prototype.hasOwnProperty,$=Array.isArray,T=console.createTask?console.createTask:function(){return null},E={},R=(u={react_stack_bottom_frame:function(e){return e()}}).react_stack_bottom_frame.bind(u,o)(),P=T(a(o)),C={};l.Fragment=b,l.jsx=function(e,r,t,n,o){var i=1e4>w.recentlyCreatedOwnerStacks++;return s(e,r,t,!1,0,o,i?Error("react-stack-top-frame"):R,i?T(a(e)):P)},l.jsxs=function(e,r,t,n,o){var i=1e4>w.recentlyCreatedOwnerStacks++;return s(e,r,t,!0,0,o,i?Error("react-stack-top-frame"):R,i?T(a(e)):P)}}()),l);var s=a.exports;const c=t.button`
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
`,d=t.header`
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--omx-spacing-3) var(--omx-spacing-4);
  background: var(--omx-color-bg-default, #fff);
  border-bottom: 1px solid rgba(0,0,0,0.06);
`,u=t.div`
  display: flex; align-items: center; gap: var(--omx-spacing-2);
  font-weight: 600; color: var(--omx-color-text-default, #0f172a);
`,p=t.nav`
  display: none; gap: var(--omx-spacing-4);
  @media (min-width: 768px) { display: flex; }
`,f=t.a`
  color: #2563eb; text-decoration: none;
  &:hover { text-decoration: underline; }
`,b=t.div`
  display: flex; align-items: center; gap: var(--omx-spacing-3);
`,m=t.button`
  background: transparent; border: none; cursor: pointer; font-size: 1.25rem;
`,y=t.aside`
  position: sticky; top: 0; height: 100dvh; z-index: 40;
  width: ${e=>e.collapsed?"64px":"240px"};
  background: var(--omx-color-bg-subtle, #0ea5e91a);
  border-right: 1px solid rgba(0,0,0,0.06);
  transition: width .2s ease;
  display: flex; flex-direction: column; gap: 4px; padding: 8px;
`,g=t.a`
  display: flex; align-items: center; gap: 8px; padding: 8px; border-radius: 8px;
  color: ${e=>e.active?"#0f172a":"#334155"}; text-decoration: none;
  background: ${e=>e.active?"rgba(14,165,233,0.12)":"transparent"};
  &:hover { background: rgba(14,165,233,0.10); }
  > span.label { display: ${e=>e.collapsed?"none":"inline"}; }
`,x=t.button`
  background: transparent; border: none; cursor: pointer; padding: 8px; text-align: left;
`,v=t.nav`
  font-size: 0.875rem; color: #475569; padding: 8px 16px;
`,h=t.ol`
  display: flex; align-items: center; gap: 8px; list-style: none; padding: 0; margin: 0;
`,j=t.a`
  color: #2563eb; text-decoration: none; &:hover{ text-decoration: underline; }
`;e.Breadcrumbs=({items:e,"aria-label":r="Breadcrumb"})=>s.jsx(v,{"aria-label":r,className:"omx-v2",children:s.jsx(h,{children:e.map((r,t)=>{const n=t===e.length-1;return s.jsxs("li",{"aria-current":n?"page":void 0,children:[r.href&&!n?s.jsx(j,{href:r.href,children:r.label}):s.jsx("span",{children:r.label}),!n&&s.jsx("span",{"aria-hidden":!0,style:{margin:"0 4px"},children:"/"})]},t)})})}),e.Button=({variant:e="primary",children:r,...t})=>s.jsx(c,{className:"omx-v2",$variant:e,...t,children:r}),e.Header=({brand:e="OllamaMax",links:r=[],onToggleMenu:t,user:n})=>s.jsxs(d,{className:"omx-v2",role:"banner",children:[s.jsxs(u,{children:[s.jsx(m,{"aria-label":"Open menu",onClick:t,style:{display:"inline-flex"},children:s.jsx("span",{"aria-hidden":!0,children:"☰"})}),s.jsx("span",{children:e})]}),s.jsx(p,{"aria-label":"Primary",children:r.map(e=>s.jsx(f,{href:e.href,children:e.label},e.href))}),s.jsxs(b,{children:[s.jsx(m,{"aria-label":"Notifications",children:s.jsx("span",{"aria-hidden":!0,children:"🔔"})}),n?.name&&s.jsxs("div",{"aria-label":"User menu",children:[n.name," ",n.onLogout&&s.jsx("button",{onClick:n.onLogout,"aria-label":"Logout",style:{marginLeft:8},children:"Logout"})]})]})]}),e.SideNav=({items:e,activeHref:t,collapsedKey:n="omx-sidenav",defaultCollapsed:a=!1,onNavigate:o})=>{const[i,l]=r.useState(a);r.useEffect(()=>{const e=localStorage.getItem(n);null!=e&&l("1"===e)},[n]);return s.jsxs(y,{className:"omx-v2",role:"navigation","aria-label":"Sidebar",collapsed:i,children:[s.jsx(x,{onClick:()=>{l(e=>{const r=!e;try{localStorage.setItem(n,r?"1":"0")}catch{}return r})},"aria-label":i?"Expand sidebar":"Collapse sidebar",children:i?"»":"«"}),s.jsx("div",{role:"list",children:e.map(e=>s.jsxs(g,{href:e.href,active:t===e.href,collapsed:i,onClick:r=>{o&&(r.preventDefault(),o(e.href))},role:"listitem",children:[e.icon&&s.jsx("span",{"aria-hidden":!0,children:e.icon}),s.jsx("span",{className:"label",children:e.label})]},e.href))})]})}});
//# sourceMappingURL=index.umd.js.map
