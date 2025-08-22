import { writeFileSync, readFileSync, mkdirSync } from 'node:fs'

const tokens = JSON.parse(readFileSync(new URL('../src/tokens.json', import.meta.url), 'utf8'))

function toCssVars(obj, prefix = '--omx') {
  let css = ''
  for (const [k, v] of Object.entries(obj)) {
    const name = `${prefix}-${k}`
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      css += toCssVars(v, name)
    } else {
      css += `  ${name}: ${v};\n`
    }
  }
  return css
}

const css = `:root {\n${toCssVars(tokens)}}\n`

// Ensure dist directory exists
const distDir = new URL('../dist', import.meta.url)
mkdirSync(distDir, { recursive: true })

writeFileSync(new URL('../dist/omx-tokens.css', import.meta.url), css, 'utf8')

