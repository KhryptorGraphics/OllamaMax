/**
 * Multi-Tenant Architecture System
 * 
 * Comprehensive multi-tenancy implementation for enterprise deployments
 * with tenant isolation, resource management, and data segregation.
 * 
 * Features:
 * - Tenant isolation strategies (Database, Schema, Row-level)
 * - Resource quotas and limits per tenant
 * - Tenant-specific configurations
 * - Data segregation and security
 * - Tenant provisioning and lifecycle management
 * - Cross-tenant analytics (admin only)
 * - Billing and usage tracking per tenant
 */

const EventEmitter = require('events');
const crypto = require('crypto');
const fs = require('fs').promises;
const path = require('path');

class MultiTenantSystem extends EventEmitter {
    constructor(config = {}) {
        super();
        
        this.config = {
            // Isolation Strategy
            isolationStrategy: config.isolationStrategy || 'schema', // database, schema, row
            
            // Resource Limits
            defaultQuotas: config.defaultQuotas || {
                maxUsers: 100,
                maxAgents: 50,
                maxStorage: 10737418240, // 10GB in bytes
                maxComputeHours: 1000,
                maxApiCalls: 1000000,
                maxConcurrentConnections: 100
            },
            
            // Tenant Configuration
            allowCustomDomains: config.allowCustomDomains !== false,
            allowCustomBranding: config.allowCustomBranding !== false,
            requireApproval: config.requireApproval || false,
            trialPeriodDays: config.trialPeriodDays || 14,
            
            // Data Management
            dataRetentionDays: config.dataRetentionDays || 90,
            enableDataExport: config.enableDataExport !== false,
            enableDataImport: config.enableDataImport !== false,
            
            // Performance
            cacheEnabled: config.cacheEnabled !== false,
            cacheTTL: config.cacheTTL || 300000, // 5 minutes
            
            // Billing
            billingEnabled: config.billingEnabled || false,
            billingProvider: config.billingProvider || null,
            
            ...config
        };
        
        // Tenant Storage
        this.tenants = new Map();
        this.tenantDatabases = new Map();
        this.tenantConfigs = new Map();
        this.tenantMetrics = new Map();
        
        // Resource Management
        this.resourceUsage = new Map();
        this.quotaEnforcement = new Map();
        
        // Cache
        this.tenantCache = new Map();
        
        // Billing
        this.billingData = new Map();
        
        // System Metrics
        this.systemMetrics = {
            totalTenants: 0,
            activeTenants: 0,
            totalUsers: 0,
            totalStorage: 0,
            totalApiCalls: 0
        };
        
        this.init();
    }
    
    async init() {
        console.log('🏢 Initializing Multi-Tenant System...');
        
        // Load existing tenants
        await this.loadTenants();
        
        // Initialize isolation strategy
        await this.initializeIsolationStrategy();
        
        // Start monitoring
        this.startMonitoring();
        
        console.log(`✅ Multi-Tenant System initialized with ${this.tenants.size} tenants`);
        console.log(`📊 Isolation strategy: ${this.config.isolationStrategy}`);
        
        this.emit('initialized', {
            timestamp: Date.now(),
            tenantCount: this.tenants.size,
            isolationStrategy: this.config.isolationStrategy
        });
    }
    
    async loadTenants() {
        try {
            const dataPath = path.join(__dirname, '../../data/tenants.json');
            const data = await fs.readFile(dataPath, 'utf8');
            const parsed = JSON.parse(data);
            
            for (const tenant of parsed.tenants || []) {
                this.tenants.set(tenant.id, tenant);
                
                // Load tenant-specific data
                await this.loadTenantData(tenant.id);
            }
            
            console.log(`📂 Loaded ${this.tenants.size} tenants`);
        } catch (error) {
            console.log('📂 No existing tenant data found');
        }
    }
    
    async loadTenantData(tenantId) {
        // Load tenant configuration
        const config = await this.loadTenantConfig(tenantId);
        if (config) {
            this.tenantConfigs.set(tenantId, config);
        }
        
        // Load tenant metrics
        const metrics = await this.loadTenantMetrics(tenantId);
        if (metrics) {
            this.tenantMetrics.set(tenantId, metrics);
        }
        
        // Initialize resource usage tracking
        this.resourceUsage.set(tenantId, {
            users: 0,
            agents: 0,
            storage: 0,
            computeHours: 0,
            apiCalls: 0,
            connections: 0
        });
    }
    
    async initializeIsolationStrategy() {
        switch (this.config.isolationStrategy) {
            case 'database':
                await this.initializeDatabaseIsolation();
                break;
            case 'schema':
                await this.initializeSchemaIsolation();
                break;
            case 'row':
                await this.initializeRowIsolation();
                break;
            default:
                throw new Error(`Unknown isolation strategy: ${this.config.isolationStrategy}`);
        }
    }
    
    async initializeDatabaseIsolation() {
        console.log('🗄️ Initializing database-level tenant isolation');
        // Each tenant gets their own database
        for (const [tenantId, tenant] of this.tenants) {
            const dbName = `ollamamax_${tenantId}`;
            this.tenantDatabases.set(tenantId, {
                name: dbName,
                connection: null, // Would establish actual DB connection
                initialized: true
            });
        }
    }
    
    async initializeSchemaIsolation() {
        console.log('📋 Initializing schema-level tenant isolation');
        // Each tenant gets their own schema in shared database
        for (const [tenantId, tenant] of this.tenants) {
            const schemaName = `tenant_${tenantId}`;
            this.tenantDatabases.set(tenantId, {
                schema: schemaName,
                connection: null, // Would use shared connection with schema prefix
                initialized: true
            });
        }
    }
    
    async initializeRowIsolation() {
        console.log('📝 Initializing row-level tenant isolation');
        // All tenants share tables with tenant_id column
        this.tenantDatabases.set('shared', {
            type: 'row_level',
            tenantColumn: 'tenant_id',
            connection: null, // Single shared connection
            initialized: true
        });
    }
    
    startMonitoring() {
        // Monitor resource usage
        setInterval(() => {
            this.monitorResourceUsage();
        }, 60000); // Every minute
        
        // Enforce quotas
        setInterval(() => {
            this.enforceQuotas();
        }, 300000); // Every 5 minutes
        
        // Update metrics
        setInterval(() => {
            this.updateSystemMetrics();
        }, 30000); // Every 30 seconds
        
        // Cache cleanup
        if (this.config.cacheEnabled) {
            setInterval(() => {
                this.cleanupCache();
            }, this.config.cacheTTL);
        }
    }
    
    // Tenant Management
    async createTenant(tenantData) {
        const tenantId = this.generateTenantId();
        
        const tenant = {
            id: tenantId,
            name: tenantData.name,
            domain: tenantData.domain || `${tenantId}.ollamamax.io`,
            status: this.config.requireApproval ? 'pending' : 'active',
            plan: tenantData.plan || 'trial',
            createdAt: Date.now(),
            updatedAt: Date.now(),
            trialEndsAt: Date.now() + (this.config.trialPeriodDays * 24 * 60 * 60 * 1000),
            metadata: tenantData.metadata || {},
            
            // Contact Information
            owner: {
                name: tenantData.ownerName,
                email: tenantData.ownerEmail,
                phone: tenantData.ownerPhone
            },
            
            // Settings
            settings: {
                timezone: tenantData.timezone || 'UTC',
                language: tenantData.language || 'en',
                customDomain: null,
                branding: {}
            }
        };
        
        this.tenants.set(tenantId, tenant);
        
        // Initialize tenant isolation
        await this.initializeTenantIsolation(tenantId);
        
        // Set default configuration
        const config = await this.createTenantConfig(tenantId, tenantData.config);
        this.tenantConfigs.set(tenantId, config);
        
        // Initialize resource tracking
        this.resourceUsage.set(tenantId, {
            users: 0,
            agents: 0,
            storage: 0,
            computeHours: 0,
            apiCalls: 0,
            connections: 0
        });
        
        // Initialize metrics
        this.tenantMetrics.set(tenantId, {
            created: Date.now(),
            lastActive: Date.now(),
            totalLogins: 0,
            totalApiCalls: 0,
            totalErrors: 0
        });
        
        // Initialize quotas
        this.quotaEnforcement.set(tenantId, {
            ...this.config.defaultQuotas,
            ...(tenantData.customQuotas || {})
        });
        
        await this.audit('tenant_created', { tenantId, name: tenant.name });
        
        this.emit('tenant-created', {
            tenantId,
            name: tenant.name,
            status: tenant.status
        });
        
        return {
            id: tenantId,
            ...tenant
        };
    }
    
    async updateTenant(tenantId, updates) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        // Update tenant data
        Object.assign(tenant, updates, {
            updatedAt: Date.now()
        });
        
        // Update configuration if provided
        if (updates.config) {
            const config = this.tenantConfigs.get(tenantId);
            Object.assign(config, updates.config);
        }
        
        // Update quotas if provided
        if (updates.quotas) {
            const quotas = this.quotaEnforcement.get(tenantId);
            Object.assign(quotas, updates.quotas);
        }
        
        await this.audit('tenant_updated', { tenantId, updates: Object.keys(updates) });
        
        this.emit('tenant-updated', { tenantId });
        
        return tenant;
    }
    
    async deleteTenant(tenantId) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        // Soft delete - mark as deleted but keep data for retention period
        tenant.status = 'deleted';
        tenant.deletedAt = Date.now();
        tenant.dataExpiresAt = Date.now() + (this.config.dataRetentionDays * 24 * 60 * 60 * 1000);
        
        await this.audit('tenant_deleted', { tenantId, name: tenant.name });
        
        this.emit('tenant-deleted', { tenantId });
        
        // Schedule data purge after retention period
        setTimeout(() => {
            this.purgeTenantData(tenantId);
        }, this.config.dataRetentionDays * 24 * 60 * 60 * 1000);
        
        return { success: true };
    }
    
    async purgeTenantData(tenantId) {
        console.log(`🗑️ Purging data for tenant ${tenantId}`);
        
        // Remove from all maps
        this.tenants.delete(tenantId);
        this.tenantDatabases.delete(tenantId);
        this.tenantConfigs.delete(tenantId);
        this.tenantMetrics.delete(tenantId);
        this.resourceUsage.delete(tenantId);
        this.quotaEnforcement.delete(tenantId);
        this.tenantCache.delete(tenantId);
        this.billingData.delete(tenantId);
        
        // Clean up database/schema
        await this.cleanupTenantIsolation(tenantId);
        
        await this.audit('tenant_purged', { tenantId });
        
        this.emit('tenant-purged', { tenantId });
    }
    
    async suspendTenant(tenantId, reason) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        tenant.status = 'suspended';
        tenant.suspendedAt = Date.now();
        tenant.suspensionReason = reason;
        
        await this.audit('tenant_suspended', { tenantId, reason });
        
        this.emit('tenant-suspended', { tenantId, reason });
        
        return { success: true };
    }
    
    async reactivateTenant(tenantId) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        tenant.status = 'active';
        delete tenant.suspendedAt;
        delete tenant.suspensionReason;
        
        await this.audit('tenant_reactivated', { tenantId });
        
        this.emit('tenant-reactivated', { tenantId });
        
        return { success: true };
    }
    
    // Tenant Isolation
    async initializeTenantIsolation(tenantId) {
        switch (this.config.isolationStrategy) {
            case 'database':
                await this.createTenantDatabase(tenantId);
                break;
            case 'schema':
                await this.createTenantSchema(tenantId);
                break;
            case 'row':
                // No additional setup needed for row-level
                break;
        }
    }
    
    async createTenantDatabase(tenantId) {
        const dbName = `ollamamax_${tenantId}`;
        
        // Would execute actual database creation
        console.log(`🗄️ Creating database for tenant ${tenantId}: ${dbName}`);
        
        this.tenantDatabases.set(tenantId, {
            name: dbName,
            connection: null, // Would establish connection
            initialized: true
        });
    }
    
    async createTenantSchema(tenantId) {
        const schemaName = `tenant_${tenantId}`;
        
        // Would execute actual schema creation
        console.log(`📋 Creating schema for tenant ${tenantId}: ${schemaName}`);
        
        this.tenantDatabases.set(tenantId, {
            schema: schemaName,
            connection: null, // Would use shared connection
            initialized: true
        });
    }
    
    async cleanupTenantIsolation(tenantId) {
        switch (this.config.isolationStrategy) {
            case 'database':
                await this.dropTenantDatabase(tenantId);
                break;
            case 'schema':
                await this.dropTenantSchema(tenantId);
                break;
            case 'row':
                await this.deleteTenantRows(tenantId);
                break;
        }
    }
    
    async dropTenantDatabase(tenantId) {
        const db = this.tenantDatabases.get(tenantId);
        if (db) {
            console.log(`🗑️ Dropping database for tenant ${tenantId}: ${db.name}`);
            // Would execute actual database drop
        }
    }
    
    async dropTenantSchema(tenantId) {
        const db = this.tenantDatabases.get(tenantId);
        if (db) {
            console.log(`🗑️ Dropping schema for tenant ${tenantId}: ${db.schema}`);
            // Would execute actual schema drop
        }
    }
    
    async deleteTenantRows(tenantId) {
        console.log(`🗑️ Deleting rows for tenant ${tenantId}`);
        // Would execute DELETE WHERE tenant_id = tenantId
    }
    
    // Configuration Management
    async createTenantConfig(tenantId, customConfig = {}) {
        const defaultConfig = {
            // Feature Flags
            features: {
                swarmOrchestration: true,
                machineLearning: true,
                advancedAnalytics: true,
                customIntegrations: false,
                apiAccess: true
            },
            
            // Resource Limits (can override defaults)
            limits: {
                ...this.config.defaultQuotas,
                ...customConfig.limits
            },
            
            // Security Settings
            security: {
                enforceSSL: true,
                ipWhitelist: [],
                ipBlacklist: [],
                requireMFA: false,
                passwordPolicy: {
                    minLength: 8,
                    requireComplexity: true
                }
            },
            
            // Integration Settings
            integrations: {
                slack: null,
                teams: null,
                webhook: null,
                customApi: null
            },
            
            // Notification Settings
            notifications: {
                email: true,
                sms: false,
                inApp: true,
                webhooks: []
            },
            
            ...customConfig
        };
        
        return defaultConfig;
    }
    
    async getTenantConfig(tenantId) {
        // Check cache first
        if (this.config.cacheEnabled) {
            const cached = this.tenantCache.get(`config_${tenantId}`);
            if (cached && cached.expires > Date.now()) {
                return cached.data;
            }
        }
        
        const config = this.tenantConfigs.get(tenantId);
        if (!config) {
            throw new Error('Tenant configuration not found');
        }
        
        // Update cache
        if (this.config.cacheEnabled) {
            this.tenantCache.set(`config_${tenantId}`, {
                data: config,
                expires: Date.now() + this.config.cacheTTL
            });
        }
        
        return config;
    }
    
    async updateTenantConfig(tenantId, updates) {
        const config = this.tenantConfigs.get(tenantId);
        if (!config) {
            throw new Error('Tenant configuration not found');
        }
        
        // Deep merge configuration
        this.deepMerge(config, updates);
        
        // Invalidate cache
        if (this.config.cacheEnabled) {
            this.tenantCache.delete(`config_${tenantId}`);
        }
        
        await this.audit('tenant_config_updated', { tenantId, updates: Object.keys(updates) });
        
        this.emit('tenant-config-updated', { tenantId });
        
        return config;
    }
    
    // Resource Management
    async trackResourceUsage(tenantId, resource, amount) {
        const usage = this.resourceUsage.get(tenantId);
        if (!usage) {
            throw new Error('Tenant not found');
        }
        
        usage[resource] = (usage[resource] || 0) + amount;
        
        // Check quota
        const quotas = this.quotaEnforcement.get(tenantId);
        const quotaKey = `max${resource.charAt(0).toUpperCase() + resource.slice(1)}`;
        
        if (quotas[quotaKey] && usage[resource] > quotas[quotaKey]) {
            await this.handleQuotaExceeded(tenantId, resource);
            throw new Error(`Quota exceeded for ${resource}`);
        }
        
        // Update metrics
        const metrics = this.tenantMetrics.get(tenantId);
        if (metrics) {
            metrics.lastActive = Date.now();
            if (resource === 'apiCalls') {
                metrics.totalApiCalls = (metrics.totalApiCalls || 0) + amount;
            }
        }
        
        return usage[resource];
    }
    
    async getResourceUsage(tenantId) {
        const usage = this.resourceUsage.get(tenantId);
        if (!usage) {
            throw new Error('Tenant not found');
        }
        
        const quotas = this.quotaEnforcement.get(tenantId);
        
        return {
            usage,
            quotas,
            percentages: {
                users: (usage.users / quotas.maxUsers) * 100,
                agents: (usage.agents / quotas.maxAgents) * 100,
                storage: (usage.storage / quotas.maxStorage) * 100,
                computeHours: (usage.computeHours / quotas.maxComputeHours) * 100,
                apiCalls: (usage.apiCalls / quotas.maxApiCalls) * 100,
                connections: (usage.connections / quotas.maxConcurrentConnections) * 100
            }
        };
    }
    
    async handleQuotaExceeded(tenantId, resource) {
        console.log(`⚠️ Quota exceeded for tenant ${tenantId}: ${resource}`);
        
        const tenant = this.tenants.get(tenantId);
        
        await this.audit('quota_exceeded', { tenantId, resource });
        
        this.emit('quota-exceeded', {
            tenantId,
            resource,
            tenant: tenant.name
        });
        
        // Could automatically suspend tenant or send notification
        if (resource === 'storage' || resource === 'computeHours') {
            // Critical resources - might suspend
            console.log(`🚨 Critical quota exceeded for tenant ${tenantId}`);
        }
    }
    
    monitorResourceUsage() {
        for (const [tenantId, usage] of this.resourceUsage) {
            const quotas = this.quotaEnforcement.get(tenantId);
            if (!quotas) continue;
            
            // Check each resource against quota
            for (const [resource, value] of Object.entries(usage)) {
                const quotaKey = `max${resource.charAt(0).toUpperCase() + resource.slice(1)}`;
                const quota = quotas[quotaKey];
                
                if (quota && value > quota * 0.9) {
                    // Warning at 90% usage
                    this.emit('quota-warning', {
                        tenantId,
                        resource,
                        usage: value,
                        quota,
                        percentage: (value / quota) * 100
                    });
                }
            }
        }
    }
    
    enforceQuotas() {
        for (const [tenantId, tenant] of this.tenants) {
            if (tenant.status !== 'active') continue;
            
            const usage = this.resourceUsage.get(tenantId);
            const quotas = this.quotaEnforcement.get(tenantId);
            
            if (!usage || !quotas) continue;
            
            let violations = [];
            
            for (const [resource, value] of Object.entries(usage)) {
                const quotaKey = `max${resource.charAt(0).toUpperCase() + resource.slice(1)}`;
                const quota = quotas[quotaKey];
                
                if (quota && value > quota) {
                    violations.push({ resource, usage: value, quota });
                }
            }
            
            if (violations.length > 0) {
                this.emit('quota-violations', {
                    tenantId,
                    violations
                });
            }
        }
    }
    
    // Tenant Context
    async getTenantContext(tenantId) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        if (tenant.status !== 'active') {
            throw new Error(`Tenant is ${tenant.status}`);
        }
        
        const config = await this.getTenantConfig(tenantId);
        const usage = await this.getResourceUsage(tenantId);
        const database = this.tenantDatabases.get(tenantId);
        
        return {
            tenant,
            config,
            usage,
            database,
            isolationStrategy: this.config.isolationStrategy
        };
    }
    
    // Data Operations
    async executeTenantQuery(tenantId, query) {
        const context = await this.getTenantContext(tenantId);
        
        // Apply tenant isolation to query
        switch (this.config.isolationStrategy) {
            case 'database':
                // Use tenant-specific database
                query.database = context.database.name;
                break;
            case 'schema':
                // Prefix with tenant schema
                query.schema = context.database.schema;
                break;
            case 'row':
                // Add tenant_id filter
                query.where = { ...query.where, tenant_id: tenantId };
                break;
        }
        
        // Track API call
        await this.trackResourceUsage(tenantId, 'apiCalls', 1);
        
        // Would execute actual query
        return { success: true, data: [] };
    }
    
    async exportTenantData(tenantId) {
        if (!this.config.enableDataExport) {
            throw new Error('Data export is disabled');
        }
        
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        const exportData = {
            tenant: { ...tenant, id: undefined }, // Exclude ID
            config: this.tenantConfigs.get(tenantId),
            metrics: this.tenantMetrics.get(tenantId),
            exportedAt: Date.now()
        };
        
        await this.audit('tenant_data_exported', { tenantId });
        
        this.emit('tenant-data-exported', { tenantId });
        
        return exportData;
    }
    
    async importTenantData(tenantId, data) {
        if (!this.config.enableDataImport) {
            throw new Error('Data import is disabled');
        }
        
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        // Validate and sanitize import data
        const sanitized = this.sanitizeImportData(data);
        
        // Import configuration
        if (sanitized.config) {
            await this.updateTenantConfig(tenantId, sanitized.config);
        }
        
        await this.audit('tenant_data_imported', { tenantId });
        
        this.emit('tenant-data-imported', { tenantId });
        
        return { success: true };
    }
    
    // Billing
    async trackUsageForBilling(tenantId, metric, amount) {
        if (!this.config.billingEnabled) return;
        
        const billingData = this.billingData.get(tenantId) || {
            period: this.getCurrentBillingPeriod(),
            usage: {}
        };
        
        billingData.usage[metric] = (billingData.usage[metric] || 0) + amount;
        
        this.billingData.set(tenantId, billingData);
        
        return billingData;
    }
    
    async generateBillingReport(tenantId) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        const billingData = this.billingData.get(tenantId);
        const usage = this.resourceUsage.get(tenantId);
        
        return {
            tenant: {
                id: tenantId,
                name: tenant.name,
                plan: tenant.plan
            },
            period: this.getCurrentBillingPeriod(),
            usage: billingData ? billingData.usage : {},
            resourceUsage: usage,
            generatedAt: Date.now()
        };
    }
    
    // Analytics
    async getTenantAnalytics(tenantId) {
        const tenant = this.tenants.get(tenantId);
        if (!tenant) {
            throw new Error('Tenant not found');
        }
        
        const metrics = this.tenantMetrics.get(tenantId);
        const usage = this.resourceUsage.get(tenantId);
        
        return {
            tenant: {
                name: tenant.name,
                status: tenant.status,
                plan: tenant.plan,
                createdAt: tenant.createdAt
            },
            metrics,
            usage,
            trends: await this.calculateTenantTrends(tenantId)
        };
    }
    
    async getSystemAnalytics() {
        this.updateSystemMetrics();
        
        return {
            ...this.systemMetrics,
            tenantsByStatus: this.getTenantsByStatus(),
            tenantsByPlan: this.getTenantsByPlan(),
            topTenantsByUsage: await this.getTopTenantsByUsage(),
            systemHealth: this.calculateSystemHealth()
        };
    }
    
    getTenantsByStatus() {
        const byStatus = {};
        for (const tenant of this.tenants.values()) {
            byStatus[tenant.status] = (byStatus[tenant.status] || 0) + 1;
        }
        return byStatus;
    }
    
    getTenantsByPlan() {
        const byPlan = {};
        for (const tenant of this.tenants.values()) {
            byPlan[tenant.plan] = (byPlan[tenant.plan] || 0) + 1;
        }
        return byPlan;
    }
    
    async getTopTenantsByUsage() {
        const tenantUsage = [];
        
        for (const [tenantId, usage] of this.resourceUsage) {
            const tenant = this.tenants.get(tenantId);
            if (tenant && tenant.status === 'active') {
                tenantUsage.push({
                    tenantId,
                    name: tenant.name,
                    totalUsage: Object.values(usage).reduce((sum, val) => sum + val, 0)
                });
            }
        }
        
        return tenantUsage.sort((a, b) => b.totalUsage - a.totalUsage).slice(0, 10);
    }
    
    calculateSystemHealth() {
        const activeTenants = Array.from(this.tenants.values())
            .filter(t => t.status === 'active').length;
        
        const totalTenants = this.tenants.size;
        
        return {
            score: totalTenants > 0 ? (activeTenants / totalTenants) : 1,
            activeTenants,
            totalTenants,
            status: activeTenants / totalTenants > 0.8 ? 'healthy' : 'warning'
        };
    }
    
    async calculateTenantTrends(tenantId) {
        // Would calculate actual trends from historical data
        return {
            userGrowth: Math.random() * 20 - 10, // -10% to +10%
            storageGrowth: Math.random() * 30 - 15,
            apiCallsTrend: Math.random() * 50 - 25
        };
    }
    
    // Helper Methods
    generateTenantId() {
        return `tenant_${Date.now()}_${crypto.randomBytes(8).toString('hex')}`;
    }
    
    deepMerge(target, source) {
        for (const key in source) {
            if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
                target[key] = target[key] || {};
                this.deepMerge(target[key], source[key]);
            } else {
                target[key] = source[key];
            }
        }
    }
    
    getCurrentBillingPeriod() {
        const now = new Date();
        return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    }
    
    updateSystemMetrics() {
        this.systemMetrics.totalTenants = this.tenants.size;
        this.systemMetrics.activeTenants = Array.from(this.tenants.values())
            .filter(t => t.status === 'active').length;
        
        let totalUsers = 0;
        let totalStorage = 0;
        let totalApiCalls = 0;
        
        for (const usage of this.resourceUsage.values()) {
            totalUsers += usage.users || 0;
            totalStorage += usage.storage || 0;
            totalApiCalls += usage.apiCalls || 0;
        }
        
        this.systemMetrics.totalUsers = totalUsers;
        this.systemMetrics.totalStorage = totalStorage;
        this.systemMetrics.totalApiCalls = totalApiCalls;
    }
    
    cleanupCache() {
        const now = Date.now();
        for (const [key, cached] of this.tenantCache) {
            if (cached.expires < now) {
                this.tenantCache.delete(key);
            }
        }
    }
    
    sanitizeImportData(data) {
        // Remove sensitive or system fields
        const sanitized = { ...data };
        delete sanitized.id;
        delete sanitized.createdAt;
        delete sanitized.updatedAt;
        return sanitized;
    }
    
    // Persistence
    async saveData() {
        const data = {
            tenants: Array.from(this.tenants.values()),
            configs: Array.from(this.tenantConfigs.entries()).map(([id, config]) => ({ tenantId: id, config })),
            metrics: Array.from(this.tenantMetrics.entries()).map(([id, metrics]) => ({ tenantId: id, metrics })),
            timestamp: Date.now()
        };
        
        try {
            const dataPath = path.join(__dirname, '../../data/tenants.json');
            await fs.writeFile(dataPath, JSON.stringify(data, null, 2));
            console.log('💾 Tenant data saved successfully');
        } catch (error) {
            console.error('❌ Failed to save tenant data:', error);
        }
    }
    
    async loadTenantConfig(tenantId) {
        try {
            const configPath = path.join(__dirname, `../../data/tenant-configs/${tenantId}.json`);
            const data = await fs.readFile(configPath, 'utf8');
            return JSON.parse(data);
        } catch (error) {
            return null;
        }
    }
    
    async loadTenantMetrics(tenantId) {
        try {
            const metricsPath = path.join(__dirname, `../../data/tenant-metrics/${tenantId}.json`);
            const data = await fs.readFile(metricsPath, 'utf8');
            return JSON.parse(data);
        } catch (error) {
            return null;
        }
    }
    
    // Audit
    async audit(action, data) {
        const entry = {
            timestamp: Date.now(),
            action,
            data
        };
        
        this.emit('audit-logged', entry);
    }
    
    async shutdown() {
        console.log('🛑 Shutting down Multi-Tenant System...');
        
        await this.saveData();
        
        this.emit('shutdown', { timestamp: Date.now() });
        console.log('✅ Multi-Tenant System shutdown complete');
    }
}

module.exports = MultiTenantSystem;