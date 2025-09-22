/**
 * User Model for Ollamamax
 * Handles user data persistence and validation
 */

const bcrypt = require('bcrypt');
const sqlite3 = require('sqlite3').verbose();
const path = require('path');
const fs = require('fs');
const crypto = require('crypto');

class UserModel {
  constructor() {
    this.dbPath = process.env.DB_PATH || path.join(__dirname, '../../data/ollamamax.db');
    this.db = null;
    this.saltRounds = 12;
    this.init();
  }

  async init() {
    return new Promise((resolve, reject) => {
      // Create data directory if it doesn't exist
      const dataDir = path.dirname(this.dbPath);
      fs.mkdirSync(dataDir, { recursive: true });

      this.db = new sqlite3.Database(this.dbPath, (err) => {
        if (err) {
          console.error('Database connection error:', err);
          reject(err);
        } else {
          console.log('Connected to SQLite database');
          this.createTables().then(resolve).catch(reject);
        }
      });
    });
  }

  async createTables() {
    const createUsersTable = `
      CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        role TEXT DEFAULT 'user',
        permissions TEXT DEFAULT '[]',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        last_login DATETIME,
        active BOOLEAN DEFAULT 1,
        email_verified BOOLEAN DEFAULT 0,
        api_key TEXT UNIQUE,
        metadata TEXT DEFAULT '{}'
      )
    `;

    const createUserSessionsTable = `
      CREATE TABLE IF NOT EXISTS user_sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        refresh_token TEXT UNIQUE,
        expires_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        ip_address TEXT,
        user_agent TEXT,
        active BOOLEAN DEFAULT 1,
        FOREIGN KEY (user_id) REFERENCES users (id)
      )
    `;

    const createUserUsageTable = `
      CREATE TABLE IF NOT EXISTS user_usage (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        date DATE,
        requests_count INTEGER DEFAULT 0,
        tokens_used INTEGER DEFAULT 0,
        compute_seconds REAL DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id),
        UNIQUE(user_id, date)
      )
    `;

    const createApiKeysTable = `
      CREATE TABLE IF NOT EXISTS api_keys (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        key_hash TEXT UNIQUE,
        name TEXT,
        permissions TEXT DEFAULT '["inference"]',
        last_used DATETIME,
        expires_at DATETIME,
        active BOOLEAN DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id)
      )
    `;

    return new Promise((resolve, reject) => {
      this.db.serialize(() => {
        this.db.run(createUsersTable);
        this.db.run(createUserSessionsTable);
        this.db.run(createUserUsageTable);
        this.db.run(createApiKeysTable, (err) => {
          if (err) reject(err);
          else resolve();
        });
      });
    });
  }

  /**
   * Hash password using bcrypt
   */
  async hashPassword(password) {
    return bcrypt.hash(password, this.saltRounds);
  }

  /**
   * Verify password against hash
   */
  async verifyPassword(password, hash) {
    return bcrypt.compare(password, hash);
  }

  /**
   * Hash password synchronously
   */
  hashPasswordSync(password) {
    return bcrypt.hashSync(password, this.saltRounds);
  }

  /**
   * Generate API key
   */
  generateApiKey() {
    return 'ollamamax-' + crypto.randomBytes(32).toString('hex');
  }

  /**
   * Create new user
   */
  async createUser(userData) {
    const { email, password, role = 'user', permissions = ['inference'] } = userData;

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      throw new Error('Invalid email format');
    }

    // Validate password strength
    if (password.length < 8) {
      throw new Error('Password must be at least 8 characters long');
    }

    // Hash password
    const passwordHash = await this.hashPassword(password);
    const apiKey = this.generateApiKey();
    const apiKeyHash = await this.hashPassword(apiKey);

    return new Promise((resolve, reject) => {
      const stmt = this.db.prepare(`
        INSERT INTO users (email, password_hash, role, permissions, api_key)
        VALUES (?, ?, ?, ?, ?)
      `);

      const self = this;
      stmt.run([
        email.toLowerCase(),
        passwordHash,
        role,
        JSON.stringify(permissions),
        apiKey
      ], function(err) {
        if (err) {
          if (err.message.includes('UNIQUE constraint failed')) {
            reject(new Error('Email already exists'));
          } else {
            reject(err);
          }
        } else {
          const userId = this.lastID;
          
          // Create API key entry
          const apiStmt = self.db.prepare(`
            INSERT INTO api_keys (user_id, key_hash, name, permissions)
            VALUES (?, ?, ?, ?)
          `);
          
          apiStmt.run([
            userId,
            apiKeyHash,
            'Default API Key',
            JSON.stringify(permissions)
          ], (apiErr) => {
            if (apiErr) {
              console.error('Error creating API key:', apiErr);
            }
            
            resolve({
              id: userId,
              email: email.toLowerCase(),
              role,
              permissions,
              api_key: apiKey,
              created_at: new Date().toISOString()
            });
          });
        }
      });
    });
  }

  /**
   * Find user by email
   */
  async findByEmail(email) {
    return new Promise((resolve, reject) => {
      this.db.get(
        'SELECT * FROM users WHERE email = ? AND active = 1',
        [email.toLowerCase()],
        (err, row) => {
          if (err) {
            reject(err);
          } else {
            if (row) {
              row.permissions = JSON.parse(row.permissions || '[]');
              row.metadata = JSON.parse(row.metadata || '{}');
            }
            resolve(row || null);
          }
        }
      );
    });
  }

  /**
   * Find user by ID
   */
  async findById(id) {
    return new Promise((resolve, reject) => {
      this.db.get(
        'SELECT * FROM users WHERE id = ? AND active = 1',
        [id],
        (err, row) => {
          if (err) {
            reject(err);
          } else {
            if (row) {
              row.permissions = JSON.parse(row.permissions || '[]');
              row.metadata = JSON.parse(row.metadata || '{}');
            }
            resolve(row || null);
          }
        }
      );
    });
  }

  /**
   * Authenticate user with email and password
   */
  async authenticate(email, password) {
    const user = await this.findByEmail(email);
    
    if (!user) {
      throw new Error('User not found');
    }

    const isValidPassword = await this.verifyPassword(password, user.password_hash);
    
    if (!isValidPassword) {
      throw new Error('Invalid password');
    }

    // Update last login
    await this.updateLastLogin(user.id);

    // Remove password hash from response
    delete user.password_hash;
    
    return user;
  }

  /**
   * Verify API key
   */
  async verifyApiKey(apiKey) {
    return new Promise((resolve, reject) => {
      // Check if it's a direct API key
      this.db.get(
        'SELECT u.* FROM users u WHERE u.api_key = ? AND u.active = 1',
        [apiKey],
        async (err, row) => {
          if (err) {
            reject(err);
          } else if (row) {
            row.permissions = JSON.parse(row.permissions || '[]');
            row.metadata = JSON.parse(row.metadata || '{}');
            delete row.password_hash;
            resolve(row);
          } else {
            // Check API keys table
            this.db.get(`
              SELECT u.*, ak.permissions as api_permissions, ak.name as api_name
              FROM api_keys ak
              JOIN users u ON ak.user_id = u.id
              WHERE ak.key_hash = ? AND ak.active = 1 AND u.active = 1
              AND (ak.expires_at IS NULL OR ak.expires_at > datetime('now'))
            `, [this.hashPasswordSync(apiKey)], (keyErr, keyRow) => {
              if (keyErr) {
                reject(keyErr);
              } else {
                if (keyRow) {
                  keyRow.permissions = JSON.parse(keyRow.api_permissions || '[]');
                  keyRow.metadata = JSON.parse(keyRow.metadata || '{}');
                  delete keyRow.password_hash;
                  delete keyRow.api_permissions;
                  
                  // Update last used
                  this.db.run(
                    'UPDATE api_keys SET last_used = datetime(\'now\') WHERE key_hash = ?',
                    [this.hashPasswordSync(apiKey)]
                  );
                }
                resolve(keyRow || null);
              }
            });
          }
        }
      );
    });
  }

  /**
   * Update user's last login timestamp
   */
  async updateLastLogin(userId) {
    return new Promise((resolve, reject) => {
      this.db.run(
        'UPDATE users SET last_login = datetime(\'now\') WHERE id = ?',
        [userId],
        function(err) {
          if (err) reject(err);
          else resolve(this.changes);
        }
      );
    });
  }

  /**
   * Create refresh token session
   */
  async createSession(userId, refreshToken, ipAddress, userAgent) {
    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + 7); // 7 days

    return new Promise((resolve, reject) => {
      const stmt = this.db.prepare(`
        INSERT INTO user_sessions (user_id, refresh_token, expires_at, ip_address, user_agent)
        VALUES (?, ?, ?, ?, ?)
      `);

      stmt.run([
        userId,
        refreshToken,
        expiresAt.toISOString(),
        ipAddress,
        userAgent
      ], function(err) {
        if (err) reject(err);
        else resolve(this.lastID);
      });
    });
  }

  /**
   * Validate refresh token
   */
  async validateRefreshToken(refreshToken) {
    return new Promise((resolve, reject) => {
      this.db.get(`
        SELECT s.*, u.email, u.role, u.permissions
        FROM user_sessions s
        JOIN users u ON s.user_id = u.id
        WHERE s.refresh_token = ? AND s.active = 1 AND s.expires_at > datetime('now')
      `, [refreshToken], (err, row) => {
        if (err) {
          reject(err);
        } else {
          if (row) {
            row.permissions = JSON.parse(row.permissions || '[]');
          }
          resolve(row || null);
        }
      });
    });
  }

  /**
   * Revoke refresh token
   */
  async revokeRefreshToken(refreshToken) {
    return new Promise((resolve, reject) => {
      this.db.run(
        'UPDATE user_sessions SET active = 0 WHERE refresh_token = ?',
        [refreshToken],
        function(err) {
          if (err) reject(err);
          else resolve(this.changes);
        }
      );
    });
  }

  /**
   * Track user usage
   */
  async trackUsage(userId, tokensUsed, computeSeconds = 0) {
    const today = new Date().toISOString().split('T')[0];

    return new Promise((resolve, reject) => {
      this.db.run(`
        INSERT INTO user_usage (user_id, date, requests_count, tokens_used, compute_seconds)
        VALUES (?, ?, 1, ?, ?)
        ON CONFLICT(user_id, date) DO UPDATE SET
          requests_count = requests_count + 1,
          tokens_used = tokens_used + ?,
          compute_seconds = compute_seconds + ?
      `, [userId, today, tokensUsed, computeSeconds, tokensUsed, computeSeconds], function(err) {
        if (err) reject(err);
        else resolve(this.changes);
      });
    });
  }

  /**
   * Get user usage statistics
   */
  async getUserUsage(userId, days = 30) {
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);

    return new Promise((resolve, reject) => {
      this.db.all(`
        SELECT date, requests_count, tokens_used, compute_seconds
        FROM user_usage
        WHERE user_id = ? AND date >= ?
        ORDER BY date DESC
      `, [userId, startDate.toISOString().split('T')[0]], (err, rows) => {
        if (err) reject(err);
        else resolve(rows);
      });
    });
  }

  /**
   * Get user quota information
   */
  async getUserQuota(userId) {
    const today = new Date().toISOString().split('T')[0];

    return new Promise((resolve, reject) => {
      this.db.get(`
        SELECT 
          COALESCE(requests_count, 0) as requests_today,
          COALESCE(tokens_used, 0) as tokens_today,
          COALESCE(compute_seconds, 0) as compute_today
        FROM user_usage
        WHERE user_id = ? AND date = ?
      `, [userId, today], async (err, row) => {
        if (err) {
          reject(err);
        } else {
          const user = await this.findById(userId);
          const metadata = user ? user.metadata : {};
          
          resolve({
            requests_today: row ? row.requests_today : 0,
            tokens_today: row ? row.tokens_today : 0,
            compute_today: row ? row.compute_today : 0,
            limits: {
              requests_per_day: metadata.requests_limit || 1000,
              tokens_per_day: metadata.tokens_limit || 100000,
              compute_per_day: metadata.compute_limit || 3600
            }
          });
        }
      });
    });
  }

  /**
   * List all users (admin only)
   */
  async listUsers(limit = 50, offset = 0) {
    return new Promise((resolve, reject) => {
      this.db.all(`
        SELECT id, email, role, created_at, last_login, active, email_verified
        FROM users
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
      `, [limit, offset], (err, rows) => {
        if (err) reject(err);
        else resolve(rows);
      });
    });
  }

  /**
   * Update user role and permissions
   */
  async updateUser(userId, updates) {
    const allowedUpdates = ['role', 'permissions', 'active', 'metadata'];
    const updateFields = [];
    const updateValues = [];

    Object.keys(updates).forEach(key => {
      if (allowedUpdates.includes(key)) {
        updateFields.push(`${key} = ?`);
        if (key === 'permissions' || key === 'metadata') {
          updateValues.push(JSON.stringify(updates[key]));
        } else {
          updateValues.push(updates[key]);
        }
      }
    });

    if (updateFields.length === 0) {
      throw new Error('No valid update fields provided');
    }

    updateFields.push('updated_at = datetime(\'now\')');
    updateValues.push(userId);

    return new Promise((resolve, reject) => {
      this.db.run(
        `UPDATE users SET ${updateFields.join(', ')} WHERE id = ?`,
        updateValues,
        function(err) {
          if (err) reject(err);
          else resolve(this.changes);
        }
      );
    });
  }

  /**
   * Close database connection
   */
  close() {
    if (this.db) {
      this.db.close((err) => {
        if (err) {
          console.error('Error closing database:', err);
        } else {
          console.log('Database connection closed');
        }
      });
    }
  }
}

// Export singleton instance
const userModel = new UserModel();
module.exports = UserModel;