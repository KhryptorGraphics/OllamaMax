import http from 'node:http'
import { readFileSync, statSync, createReadStream, existsSync } from 'node:fs'
import { join, resolve, extname } from 'node:path'

const PORT = process.env.LEGACY_PORT ? Number(process.env.LEGACY_PORT) : 8090
const ROOT = resolve(new URL('../', import.meta.url).pathname)

const types = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.mjs': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
}

function send(res, status, headers, body) {
  res.writeHead(status, headers)
  res.end(body)
}

const server = http.createServer((req, res) => {
  try {
    const url = new URL(req.url, `http://${req.headers.host}`)
    let filePath = decodeURIComponent(url.pathname)
    if (filePath === '/') filePath = '/index.html'
    const absPath = join(ROOT, filePath)

    if (!existsSync(absPath)) {
      send(res, 404, { 'Content-Type': 'text/plain' }, 'Not Found')
      return
    }

    const st = statSync(absPath)
    if (st.isDirectory()) {
      const indexFile = join(absPath, 'index.html')
      if (existsSync(indexFile)) {
        res.writeHead(200, { 'Content-Type': types['.html'] })
        createReadStream(indexFile).pipe(res)
      } else {
        send(res, 403, { 'Content-Type': 'text/plain' }, 'Forbidden')
      }
      return
    }

    const ext = extname(absPath)
    const ctype = types[ext] || 'application/octet-stream'
    res.writeHead(200, { 'Content-Type': ctype })
    createReadStream(absPath).pipe(res)
  } catch (e) {
    send(res, 500, { 'Content-Type': 'text/plain' }, 'Internal Server Error')
  }
})

server.listen(PORT, () => {
  console.log(`[legacy] serving ${ROOT} on http://localhost:${PORT}`)
})

