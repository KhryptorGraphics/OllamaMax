#!/usr/bin/env node
/**
 * Sync OpenAPI Specification Script
 * Converts OpenAPI JSON to YAML and ensures docs/api/openapi.yaml stays in sync
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

// Minimal YAML stringifier (avoids external dependencies)
function jsonToYaml(obj, indent = 0) {
  const spaces = ' '.repeat(indent);
  let yaml = '';

  if (Array.isArray(obj)) {
    obj.forEach(item => {
      if (typeof item === 'object' && item !== null && !Array.isArray(item)) {
        yaml += `${spaces}-\n${jsonToYaml(item, indent + 2)}`;
      } else {
        yaml += `${spaces}- ${JSON.stringify(item)}\n`;
      }
    });
  } else if (typeof obj === 'object' && obj !== null) {
    Object.keys(obj).forEach(key => {
      const value = obj[key];
      if (value === null) {
        yaml += `${spaces}${key}: null\n`;
      } else if (typeof value === 'object') {
        if (Array.isArray(value) && value.length === 0) {
          yaml += `${spaces}${key}: []\n`;
        } else if (Object.keys(value).length === 0) {
          yaml += `${spaces}${key}: {}\n`;
        } else {
          yaml += `${spaces}${key}:\n${jsonToYaml(value, indent + 2)}`;
        }
      } else if (typeof value === 'string') {
        // Handle multi-line strings
        if (value.includes('\n')) {
          yaml += `${spaces}${key}: |\n`;
          value.split('\n').forEach(line => {
            yaml += `${spaces}  ${line}\n`;
          });
        } else {
          yaml += `${spaces}${key}: ${JSON.stringify(value)}\n`;
        }
      } else {
        yaml += `${spaces}${key}: ${value}\n`;
      }
    });
  }

  return yaml;
}

async function fetchOpenAPIJSON(port = 13000) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'localhost',
      port: port,
      path: '/openapi.json',
      method: 'GET',
      timeout: 5000
    };

    const req = http.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        if (res.statusCode === 200) {
          try {
            resolve(JSON.parse(data));
          } catch (e) {
            reject(new Error(`Failed to parse JSON: ${e.message}`));
          }
        } else {
          reject(new Error(`HTTP ${res.statusCode}: ${data}`));
        }
      });
    });

    req.on('error', (e) => {
      reject(new Error(`Request failed: ${e.message}`));
    });

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('Request timeout - is the server running on port 13000?'));
    });

    req.end();
  });
}

async function syncOpenAPISpec() {
  console.log('🔄 Syncing OpenAPI specification...\n');

  try {
    // Fetch JSON spec from running server
    console.log('📡 Fetching OpenAPI JSON from http://localhost:13000/openapi.json...');
    const spec = await fetchOpenAPIJSON();
    console.log('✅ OpenAPI JSON fetched successfully\n');

    // Ensure docs/api directory exists
    const docsDir = path.join(process.cwd(), 'docs', 'api');
    if (!fs.existsSync(docsDir)) {
      fs.mkdirSync(docsDir, { recursive: true });
      console.log(`📁 Created directory: ${docsDir}\n`);
    }

    // Write JSON spec
    const jsonPath = path.join(docsDir, 'openapi.json');
    fs.writeFileSync(jsonPath, JSON.stringify(spec, null, 2));
    console.log(`✅ Wrote OpenAPI JSON: ${jsonPath}\n`);

    // Convert to YAML
    console.log('🔄 Converting JSON to YAML...');
    const yaml = jsonToYaml(spec);

    // Write YAML spec
    const yamlPath = path.join(docsDir, 'openapi.yaml');
    fs.writeFileSync(yamlPath, yaml);
    console.log(`✅ Wrote OpenAPI YAML: ${yamlPath}\n`);

    // Calculate file sizes
    const jsonSize = fs.statSync(jsonPath).size;
    const yamlSize = fs.statSync(yamlPath).size;

    console.log('📊 Sync Summary:');
    console.log(`   - OpenAPI Version: ${spec.openapi}`);
    console.log(`   - API Title: ${spec.info.title}`);
    console.log(`   - API Version: ${spec.info.version}`);
    console.log(`   - Total Paths: ${Object.keys(spec.paths || {}).length}`);
    console.log(`   - JSON Size: ${(jsonSize / 1024).toFixed(2)} KB`);
    console.log(`   - YAML Size: ${(yamlSize / 1024).toFixed(2)} KB`);
    console.log('\n✨ OpenAPI specification synced successfully!');
    console.log('\n📚 Access documentation at:');
    console.log('   - Swagger UI: http://localhost:13000/docs');
    console.log('   - JSON Spec: http://localhost:13000/openapi.json');
    console.log(`   - YAML File: ${yamlPath}`);
  } catch (error) {
    console.error('\n❌ Error syncing OpenAPI specification:');
    console.error(`   ${error.message}\n`);

    if (error.message.includes('ECONNREFUSED') || error.message.includes('timeout')) {
      console.log('💡 Tip: Make sure the API server is running:');
      console.log('   npm start\n');
    }

    process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  syncOpenAPISpec();
}

module.exports = { syncOpenAPISpec, jsonToYaml };
