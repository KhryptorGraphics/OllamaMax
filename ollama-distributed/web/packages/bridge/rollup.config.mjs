import resolve from '@rollup/plugin-node-resolve'
import commonjs from '@rollup/plugin-commonjs'
import terser from '@rollup/plugin-terser'
import typescript from '@rollup/plugin-typescript'

export default {
  input: 'src/index.ts',
  output: [
    { file: 'dist/index.esm.js', format: 'esm', sourcemap: true },
    { file: 'dist/index.umd.js', format: 'umd', name: 'OmxBridge', sourcemap: true },
  ],
  external: [],
  plugins: [resolve(), commonjs(), typescript(), terser()],
}

