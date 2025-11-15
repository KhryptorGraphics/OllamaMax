/**
 * Authentication and Email Verification System for OllamaMax
 */

const sqlite3 = require('sqlite3').verbose();
const bcrypt = require('bcrypt');
const jwt = require('jsonwebtoken');
const nodemailer = require('nodemailer');
const crypto = require('crypto');
const path = require('path');

// Input sanitization utilities (ISSUE-004-XSS)
class InputSanitizer {
    static sanitizeString(input) {
        if (typeof input !== 'string') return '';
        
        // Remove null bytes and control characters
        let sanitized = input.replace(/[\x00-\x1F\x7F]/g, '');
        
        // Escape HTML entities
        sanitized = sanitized
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#x27;')
            .replace(/\//g, '&#x2F;');
        
        return sanitized;
    }
    
    static validateEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email) && email.length <= 254;
    }
    
    static validateUsername(username) {
        // Allow alphanumeric, underscore, dash, min 3 chars, max 50 chars
        const usernameRegex = /^[a-zA-Z0-9_-]{3,50}$/;
        return usernameRegex.test(username);
    }
    
    static validatePassword(password) {
        // At least 8 characters, 1 uppercase, 1 lowercase, 1 number, 1 special char
        const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/;
        return passwordRegex.test(password);
    }
}

class AuthSystem {
    constructor() {
        this.dbPath = path.join(__dirname, 'users.db');
        this.db = new sqlite3.Database(this.dbPath);

        // SECURITY: JWT_SECRET must be provided via environment variable
        this.jwtSecret = process.env.JWT_SECRET;
        if (!this.jwtSecret) {
            throw new Error('JWT_SECRET environment variable is required for security');
        }

        this.emailConfig = null;
        this.transporter = null;

        this.initializeDatabase();
        this.setupEmailTransporter();
    }

    initializeDatabase() {
        const createUsersTable = `
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT UNIQUE NOT NULL,
                email TEXT UNIQUE NOT NULL,
                password_hash TEXT NOT NULL,
                email_verified BOOLEAN DEFAULT FALSE,
                verification_token TEXT,
                verification_expires DATETIME,
                reset_token TEXT,
                reset_expires DATETIME,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                last_login DATETIME,
                is_active BOOLEAN DEFAULT TRUE
            )
        `;

        const createSessionsTable = `
            CREATE TABLE IF NOT EXISTS sessions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                token TEXT UNIQUE NOT NULL,
                expires DATETIME NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users (id)
            )
        `;

        this.db.run(createUsersTable);
        this.db.run(createSessionsTable);
        
        console.log('✅ Authentication database initialized');
    }

    async setupEmailTransporter() {
        // SECURITY: SMTP credentials must be provided via environment variables
        const smtpUser = process.env.SMTP_USER || 'noreply@giggatek.com';
        const smtpPassword = process.env.SMTP_PASSWORD;

        if (!smtpPassword) {
            console.warn('⚠️  SMTP_PASSWORD not set. Email functionality will use mock transporter.');
        }

        console.log('🔧 Testing email server connections...');
        
        for (const emailConfig of emailConfigs) {
            try {
                console.log(`  Testing ${emailConfig.name}...`);
                const transporter = nodemailer.createTransport(emailConfig.config);
                
                // Verify connection
                await transporter.verify();
                
                console.log(`  ✅ ${emailConfig.name} connection successful!`);
                this.transporter = transporter;
                this.emailConfig = emailConfig;
                
                // Send test email to verify functionality
                await this.sendTestEmail('khryptorgraphics@gmail.com');
                break;
                
            } catch (error) {
                console.log(`  ❌ ${emailConfig.name} failed: ${error.message}`);
            }
        }

        if (!this.transporter) {
            console.error('❌ All email configurations failed. Email verification disabled.');
            // Create a mock transporter for development
            this.transporter = {
                sendMail: (mailOptions) => {
                    console.log('📧 MOCK EMAIL:', {
                        to: mailOptions.to,
                        subject: mailOptions.subject,
                        text: mailOptions.text
                    });
                    return Promise.resolve({ messageId: 'mock-' + Date.now() });
                }
            };
        }
    }

    async sendTestEmail(testEmail) {
        const mailOptions = {
            from: 'noreply@giggatek.com',
            to: testEmail,
            subject: 'OllamaMax Email System Test',
            html: `
                <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
                    <h2 style="color: #2563eb;">🚀 OllamaMax Email System</h2>
                    <p>This is a test email from the OllamaMax authentication system.</p>
                    <p><strong>Configuration:</strong> ${this.emailConfig.name}</p>
                    <p><strong>Time:</strong> ${new Date().toISOString()}</p>
                    <hr>
                    <p style="color: #6b7280; font-size: 12px;">
                        This email confirms that the email verification system is working correctly.
                    </p>
                </div>
            `
        };

        try {
            const result = await this.transporter.sendMail(mailOptions);
            console.log(`  📧 Test email sent successfully to ${testEmail}`);
            console.log(`  📧 Message ID: ${result.messageId}`);
            return result;
        } catch (error) {
            console.error(`  ❌ Test email failed: ${error.message}`);
            throw error;
        }
    }

    generateToken() {
        return crypto.randomBytes(32).toString('hex');
    }

    async hashPassword(password) {
        return await bcrypt.hash(password, 10);
    }

    async verifyPassword(password, hash) {
        return await bcrypt.compare(password, hash);
    }

    generateJWT(userId) {
        return jwt.sign({ userId }, this.jwtSecret, { expiresIn: '24h' });
    }

    verifyJWT(token) {
        try {
            return jwt.verify(token, this.jwtSecret);
        } catch (error) {
            return null;
        }
    }

    async registerUser(username, email, password) {
        return new Promise((resolve, reject) => {
            // Input validation and sanitization (ISSUE-004-XSS)
            const sanitizedUsername = InputSanitizer.sanitizeString(username);
            const sanitizedEmail = email.toLowerCase().trim();
            
            if (!InputSanitizer.validateUsername(sanitizedUsername)) {
                return reject(new Error('Invalid username format. Use 3-50 characters, alphanumeric, underscore, or dash.'));
            }
            
            if (!InputSanitizer.validateEmail(sanitizedEmail)) {
                return reject(new Error('Invalid email format.'));
            }
            
            if (!InputSanitizer.validatePassword(password)) {
                return reject(new Error('Password must be at least 8 characters with uppercase, lowercase, number, and special character.'));
            }

            this.db.serialize(() => {
                // Check if user already exists
                this.db.get(
                    'SELECT id FROM users WHERE email = ? OR username = ?',
                    [sanitizedEmail, sanitizedUsername],
                    async (err, row) => {
                        if (err) return reject(err);
                        if (row) return reject(new Error('User already exists'));

                        try {
                            const passwordHash = await this.hashPassword(password);
                            const verificationToken = this.generateToken();
                            const verificationExpires = new Date(Date.now() + 24 * 60 * 60 * 1000); // 24 hours

                            this.db.run(
                                `INSERT INTO users (username, email, password_hash, verification_token, verification_expires)
                                 VALUES (?, ?, ?, ?, ?)`,
                                [sanitizedUsername, sanitizedEmail, passwordHash, verificationToken, verificationExpires],
                                function(err) {
                                    if (err) return reject(err);
                                    
                                    const userId = this.lastID;
                                    
                                    // Send verification email
                                    resolve({
                                        userId,
                                        verificationToken,
                                        message: 'User registered successfully. Please check your email for verification.'
                                    });
                                }
                            );
                        } catch (error) {
                            reject(error);
                        }
                    }
                );
            });
        });
    }

    async sendVerificationEmail(email, verificationToken, username = '') {
        const verificationUrl = `http://localhost:13100/api/verify-email?token=${verificationToken}`;
        
        const mailOptions = {
            from: 'noreply@giggatek.com',
            to: email,
            subject: 'Verify your OllamaMax account',
            html: `
                <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
                    <div style="text-align: center; margin-bottom: 30px;">
                        <h1 style="color: #2563eb; margin: 0;">🦙 OllamaMax</h1>
                        <p style="color: #6b7280; margin: 5px 0;">Distributed AI Platform</p>
                    </div>
                    
                    <div style="background: #f8fafc; padding: 30px; border-radius: 10px; border-left: 4px solid #2563eb;">
                        <h2 style="color: #1f2937; margin-top: 0;">Welcome${username ? ', ' + username : ''}! 🎉</h2>
                        <p style="color: #4b5563; line-height: 1.6;">
                            Thank you for registering with OllamaMax. To activate your account and start using our distributed AI platform, please verify your email address.
                        </p>
                        
                        <div style="text-align: center; margin: 30px 0;">
                            <a href="${verificationUrl}" 
                               style="background: #2563eb; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block;">
                                ✅ Verify Email Address
                            </a>
                        </div>
                        
                        <p style="color: #6b7280; font-size: 14px; line-height: 1.5;">
                            If the button doesn't work, copy and paste this link into your browser:
                            <br><a href="${verificationUrl}" style="color: #2563eb; word-break: break-all;">${verificationUrl}</a>
                        </p>
                        
                        <p style="color: #dc2626; font-size: 13px; margin-top: 20px;">
                            ⚠️ This verification link expires in 24 hours.
                        </p>
                    </div>
                    
                    <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb;">
                        <p style="color: #9ca3af; font-size: 12px; margin: 0;">
                            This email was sent from the OllamaMax authentication system.<br>
                            If you didn't create an account, please ignore this email.
                        </p>
                    </div>
                </div>
            `
        };

        try {
            const result = await this.transporter.sendMail(mailOptions);
            console.log(`📧 Verification email sent to ${email} (${result.messageId})`);
            return result;
        } catch (error) {
            console.error(`❌ Failed to send verification email: ${error.message}`);
            throw error;
        }
    }

    async verifyEmail(token) {
        return new Promise((resolve, reject) => {
            this.db.get(
                'SELECT * FROM users WHERE verification_token = ? AND verification_expires > CURRENT_TIMESTAMP',
                [token],
                (err, user) => {
                    if (err) return reject(err);
                    if (!user) return reject(new Error('Invalid or expired verification token'));

                    this.db.run(
                        'UPDATE users SET email_verified = TRUE, verification_token = NULL, verification_expires = NULL WHERE id = ?',
                        [user.id],
                        (err) => {
                            if (err) return reject(err);
                            resolve({
                                userId: user.id,
                                email: user.email,
                                message: 'Email verified successfully!'
                            });
                        }
                    );
                }
            );
        });
    }

    async loginUser(email, password) {
        return new Promise((resolve, reject) => {
            this.db.get(
                'SELECT * FROM users WHERE email = ? AND is_active = TRUE',
                [email],
                async (err, user) => {
                    if (err) return reject(err);
                    if (!user) return reject(new Error('Invalid email or password'));
                    if (!user.email_verified) return reject(new Error('Please verify your email address before logging in'));

                    try {
                        const validPassword = await this.verifyPassword(password, user.password_hash);
                        if (!validPassword) return reject(new Error('Invalid email or password'));

                        const token = this.generateJWT(user.id);
                        const expires = new Date(Date.now() + 24 * 60 * 60 * 1000);

                        // Store session
                        this.db.run(
                            'INSERT INTO sessions (user_id, token, expires) VALUES (?, ?, ?)',
                            [user.id, token, expires]
                        );

                        // Update last login
                        this.db.run(
                            'UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?',
                            [user.id]
                        );

                        resolve({
                            token,
                            user: {
                                id: user.id,
                                username: user.username,
                                email: user.email,
                                emailVerified: user.email_verified
                            }
                        });
                    } catch (error) {
                        reject(error);
                    }
                }
            );
        });
    }

    async validateSession(token) {
        const decoded = this.verifyJWT(token);
        if (!decoded) return null;

        return new Promise((resolve) => {
            this.db.get(
                `SELECT u.id, u.username, u.email, u.email_verified 
                 FROM users u 
                 JOIN sessions s ON u.id = s.user_id 
                 WHERE s.token = ? AND s.expires > CURRENT_TIMESTAMP AND u.is_active = TRUE`,
                [token],
                (err, user) => {
                    if (err || !user) return resolve(null);
                    resolve({
                        id: user.id,
                        username: user.username,
                        email: user.email,
                        emailVerified: user.email_verified
                    });
                }
            );
        });
    }

    async getAllUsers() {
        return new Promise((resolve, reject) => {
            this.db.all(
                'SELECT id, username, email, email_verified, created_at, last_login, is_active FROM users ORDER BY created_at DESC',
                [],
                (err, users) => {
                    if (err) return reject(err);
                    resolve(users);
                }
            );
        });
    }

    close() {
        this.db.close();
    }
}

module.exports = AuthSystem;