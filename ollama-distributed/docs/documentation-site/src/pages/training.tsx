import React, { useState } from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import clsx from 'clsx';
import styles from './training.module.css';

interface TrainingModule {
  id: string;
  title: string;
  description: string;
  duration: string;
  level: 'Beginner' | 'Intermediate' | 'Advanced';
  type: 'Tutorial' | 'Video' | 'Workshop' | 'Assessment';
  prerequisites?: string[];
  learningObjectives: string[];
  completed?: boolean;
}

const TRAINING_MODULES: TrainingModule[] = [
  // Beginner Level
  {
    id: 'getting-started',
    title: 'Getting Started with Ollama Distributed',
    description: 'Learn the basics of Ollama Distributed, including installation, configuration, and creating your first cluster.',
    duration: '45 minutes',
    level: 'Beginner',
    type: 'Tutorial',
    learningObjectives: [
      'Install Ollama Distributed on your system',
      'Configure your first node',
      'Create a basic cluster',
      'Deploy your first AI model',
      'Perform basic inference requests'
    ]
  },
  {
    id: 'web-interface-tour',
    title: 'Web Interface Guided Tour',
    description: 'Comprehensive walkthrough of the web-based control panel and monitoring dashboard.',
    duration: '30 minutes',
    level: 'Beginner',
    type: 'Video',
    learningObjectives: [
      'Navigate the web interface effectively',
      'Monitor cluster health and performance',
      'Manage models through the UI',
      'Configure basic settings',
      'Interpret system metrics and alerts'
    ]
  },
  {
    id: 'cli-fundamentals',
    title: 'Command Line Interface Fundamentals',
    description: 'Master the essential CLI commands for managing your Ollama Distributed cluster.',
    duration: '60 minutes',
    level: 'Beginner',
    type: 'Tutorial',
    learningObjectives: [
      'Use CLI for node management',
      'Execute model operations via command line',
      'Monitor cluster status and health',
      'Troubleshoot common issues',
      'Automate basic operations with scripts'
    ]
  },

  // Intermediate Level
  {
    id: 'advanced-scaling',
    title: 'Advanced Scaling Strategies',
    description: 'Learn horizontal and vertical scaling techniques, auto-scaling configuration, and performance optimization.',
    duration: '90 minutes',
    level: 'Intermediate',
    type: 'Workshop',
    prerequisites: ['getting-started', 'cli-fundamentals'],
    learningObjectives: [
      'Implement horizontal scaling strategies',
      'Configure auto-scaling based on metrics',
      'Optimize resource allocation',
      'Monitor and tune performance',
      'Handle scaling challenges and edge cases'
    ]
  },
  {
    id: 'production-deployment',
    title: 'Production Deployment Best Practices',
    description: 'Deploy Ollama Distributed in production environments using containers, orchestration, and cloud platforms.',
    duration: '120 minutes',
    level: 'Intermediate',
    type: 'Workshop',
    prerequisites: ['getting-started', 'advanced-scaling'],
    learningObjectives: [
      'Deploy with Docker and Kubernetes',
      'Configure high availability',
      'Implement security best practices',
      'Set up monitoring and alerting',
      'Plan disaster recovery strategies'
    ]
  },
  {
    id: 'api-integration',
    title: 'API Integration and SDK Usage',
    description: 'Integrate Ollama Distributed into your applications using REST APIs, WebSockets, and official SDKs.',
    duration: '75 minutes',
    level: 'Intermediate',
    type: 'Tutorial',
    prerequisites: ['getting-started'],
    learningObjectives: [
      'Use REST API for model inference',
      'Implement real-time updates with WebSockets',
      'Integrate using Go, Python, or JavaScript SDKs',
      'Handle authentication and rate limiting',
      'Build resilient client applications'
    ]
  },

  // Advanced Level
  {
    id: 'plugin-development',
    title: 'Plugin Development Workshop',
    description: 'Create custom plugins to extend Ollama Distributed functionality for your specific use cases.',
    duration: '180 minutes',
    level: 'Advanced',
    type: 'Workshop',
    prerequisites: ['api-integration', 'production-deployment'],
    learningObjectives: [
      'Understand the plugin architecture',
      'Create custom model processing plugins',
      'Develop monitoring and metrics plugins',
      'Implement authentication plugins',
      'Deploy and manage plugins in production'
    ]
  },
  {
    id: 'performance-tuning',
    title: 'Advanced Performance Tuning',
    description: 'Deep dive into performance optimization, profiling, and troubleshooting complex performance issues.',
    duration: '150 minutes',
    level: 'Advanced',
    type: 'Workshop',
    prerequisites: ['production-deployment', 'advanced-scaling'],
    learningObjectives: [
      'Profile system performance bottlenecks',
      'Optimize memory and CPU usage',
      'Tune network and storage performance',
      'Implement advanced caching strategies',
      'Troubleshoot complex performance issues'
    ]
  },
  {
    id: 'security-hardening',
    title: 'Security Hardening and Compliance',
    description: 'Implement enterprise-grade security measures and achieve compliance with industry standards.',
    duration: '120 minutes',
    level: 'Advanced',
    type: 'Workshop',
    prerequisites: ['production-deployment'],
    learningObjectives: [
      'Implement zero-trust security architecture',
      'Configure advanced authentication and authorization',
      'Set up security monitoring and alerting',
      'Achieve compliance with SOC 2, HIPAA, GDPR',
      'Perform security audits and penetration testing'
    ]
  }
];

const CERTIFICATIONS = [
  {
    title: 'Ollama Distributed User Certification',
    description: 'Demonstrates proficiency in using Ollama Distributed for AI model inference and basic cluster management.',
    level: 'Beginner',
    duration: '2-4 hours',
    prerequisites: ['getting-started', 'web-interface-tour', 'cli-fundamentals'],
    badge: 'https://img.shields.io/badge/Certified-User-blue'
  },
  {
    title: 'Ollama Distributed Developer Certification',
    description: 'Validates skills in integrating Ollama Distributed into applications and developing custom solutions.',
    level: 'Intermediate',
    duration: '4-6 hours',
    prerequisites: ['api-integration', 'production-deployment'],
    badge: 'https://img.shields.io/badge/Certified-Developer-green'
  },
  {
    title: 'Ollama Distributed Operations Certification',
    description: 'Certifies expertise in deploying, scaling, and maintaining Ollama Distributed in production environments.',
    level: 'Advanced',
    duration: '6-8 hours',
    prerequisites: ['production-deployment', 'performance-tuning', 'security-hardening'],
    badge: 'https://img.shields.io/badge/Certified-Operations-red'
  }
];

export default function Training(): JSX.Element {
  const [selectedLevel, setSelectedLevel] = useState<string>('All');
  const [selectedType, setSelectedType] = useState<string>('All');
  const [completedModules, setCompletedModules] = useState<Set<string>>(new Set());

  const filteredModules = TRAINING_MODULES.filter(module => {
    const levelMatch = selectedLevel === 'All' || module.level === selectedLevel;
    const typeMatch = selectedType === 'All' || module.type === selectedType;
    return levelMatch && typeMatch;
  });

  const toggleCompletion = (moduleId: string) => {
    const newCompleted = new Set(completedModules);
    if (newCompleted.has(moduleId)) {
      newCompleted.delete(moduleId);
    } else {
      newCompleted.add(moduleId);
    }
    setCompletedModules(newCompleted);
  };

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'Beginner': return styles.beginner;
      case 'Intermediate': return styles.intermediate;
      case 'Advanced': return styles.advanced;
      default: return '';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'Tutorial': return '📖';
      case 'Video': return '🎥';
      case 'Workshop': return '🛠️';
      case 'Assessment': return '✅';
      default: return '📚';
    }
  };

  return (
    <Layout
      title="Training Materials"
      description="Comprehensive training materials, tutorials, and certification programs for Ollama Distributed"
    >
      <div className={clsx('container', styles.training)}>
        <div className={styles.header}>
          <h1>Training & Certification</h1>
          <p>
            Master Ollama Distributed with our comprehensive training materials, 
            interactive tutorials, and professional certification programs.
          </p>
        </div>

        {/* Learning Path Overview */}
        <div className={styles.learningPath}>
          <h2>Learning Path</h2>
          <div className={styles.pathSteps}>
            <div className={clsx(styles.pathStep, styles.beginner)}>
              <h3>1. Beginner</h3>
              <p>Get started with the basics</p>
              <ul>
                <li>Installation & Setup</li>
                <li>Web Interface Tour</li>
                <li>CLI Fundamentals</li>
              </ul>
            </div>
            <div className={clsx(styles.pathStep, styles.intermediate)}>
              <h3>2. Intermediate</h3>
              <p>Learn production deployment</p>
              <ul>
                <li>Scaling Strategies</li>
                <li>Production Deployment</li>
                <li>API Integration</li>
              </ul>
            </div>
            <div className={clsx(styles.pathStep, styles.advanced)}>
              <h3>3. Advanced</h3>
              <p>Master advanced topics</p>
              <ul>
                <li>Plugin Development</li>
                <li>Performance Tuning</li>
                <li>Security Hardening</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Filters */}
        <div className={styles.filters}>
          <div className={styles.filterGroup}>
            <label>Level:</label>
            <select value={selectedLevel} onChange={(e) => setSelectedLevel(e.target.value)}>
              <option value="All">All Levels</option>
              <option value="Beginner">Beginner</option>
              <option value="Intermediate">Intermediate</option>
              <option value="Advanced">Advanced</option>
            </select>
          </div>
          <div className={styles.filterGroup}>
            <label>Type:</label>
            <select value={selectedType} onChange={(e) => setSelectedType(e.target.value)}>
              <option value="All">All Types</option>
              <option value="Tutorial">Tutorials</option>
              <option value="Video">Videos</option>
              <option value="Workshop">Workshops</option>
              <option value="Assessment">Assessments</option>
            </select>
          </div>
        </div>

        {/* Training Modules */}
        <div className={styles.modules}>
          <h2>Training Modules</h2>
          <div className={styles.moduleGrid}>
            {filteredModules.map((module) => (
              <div
                key={module.id}
                className={clsx(
                  styles.moduleCard,
                  completedModules.has(module.id) && styles.completed
                )}
              >
                <div className={styles.moduleHeader}>
                  <div className={styles.moduleIcon}>
                    {getTypeIcon(module.type)}
                  </div>
                  <div className={styles.moduleInfo}>
                    <h3>{module.title}</h3>
                    <div className={styles.moduleMeta}>
                      <span className={clsx(styles.levelBadge, getLevelColor(module.level))}>
                        {module.level}
                      </span>
                      <span className={styles.typeBadge}>{module.type}</span>
                      <span className={styles.duration}>{module.duration}</span>
                    </div>
                  </div>
                </div>

                <p className={styles.moduleDescription}>{module.description}</p>

                {module.prerequisites && (
                  <div className={styles.prerequisites}>
                    <strong>Prerequisites:</strong>
                    <ul>
                      {module.prerequisites.map((prereq) => (
                        <li key={prereq}>
                          {TRAINING_MODULES.find(m => m.id === prereq)?.title || prereq}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className={styles.learningObjectives}>
                  <strong>Learning Objectives:</strong>
                  <ul>
                    {module.learningObjectives.map((objective, index) => (
                      <li key={index}>{objective}</li>
                    ))}
                  </ul>
                </div>

                <div className={styles.moduleActions}>
                  <Link
                    to={`/docs/training/${module.id}`}
                    className={styles.startButton}
                  >
                    {completedModules.has(module.id) ? 'Review' : 'Start Module'}
                  </Link>
                  <button
                    className={styles.completeButton}
                    onClick={() => toggleCompletion(module.id)}
                  >
                    {completedModules.has(module.id) ? '✅ Completed' : 'Mark Complete'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Certifications */}
        <div className={styles.certifications}>
          <h2>Professional Certifications</h2>
          <div className={styles.certificationGrid}>
            {CERTIFICATIONS.map((cert, index) => (
              <div key={index} className={styles.certificationCard}>
                <div className={styles.certHeader}>
                  <h3>{cert.title}</h3>
                  <img src={cert.badge} alt={`${cert.title} Badge`} />
                </div>
                <p>{cert.description}</p>
                <div className={styles.certMeta}>
                  <span className={clsx(styles.levelBadge, getLevelColor(cert.level))}>
                    {cert.level}
                  </span>
                  <span className={styles.duration}>Duration: {cert.duration}</span>
                </div>
                <div className={styles.prerequisites}>
                  <strong>Prerequisites:</strong>
                  <ul>
                    {cert.prerequisites.map((prereq) => (
                      <li key={prereq}>
                        {TRAINING_MODULES.find(m => m.id === prereq)?.title || prereq}
                      </li>
                    ))}
                  </ul>
                </div>
                <Link
                  to={`/docs/training/certification-${cert.level.toLowerCase()}`}
                  className={styles.certButton}
                >
                  View Certification Details
                </Link>
              </div>
            ))}
          </div>
        </div>

        {/* Progress Tracking */}
        <div className={styles.progress}>
          <h2>Your Progress</h2>
          <div className={styles.progressStats}>
            <div className={styles.progressStat}>
              <span className={styles.statNumber}>{completedModules.size}</span>
              <span className={styles.statLabel}>Modules Completed</span>
            </div>
            <div className={styles.progressStat}>
              <span className={styles.statNumber}>
                {Math.round((completedModules.size / TRAINING_MODULES.length) * 100)}%
              </span>
              <span className={styles.statLabel}>Overall Progress</span>
            </div>
            <div className={styles.progressStat}>
              <span className={styles.statNumber}>
                {CERTIFICATIONS.filter(cert => 
                  cert.prerequisites.every(prereq => completedModules.has(prereq))
                ).length}
              </span>
              <span className={styles.statLabel}>Certifications Ready</span>
            </div>
          </div>
        </div>

        {/* Getting Help */}
        <div className={styles.help}>
          <h2>Need Help?</h2>
          <div className={styles.helpGrid}>
            <div className={styles.helpCard}>
              <h3>📚 Documentation</h3>
              <p>Comprehensive guides and API references</p>
              <Link to="/docs">Browse Documentation</Link>
            </div>
            <div className={styles.helpCard}>
              <h3>💬 Community Support</h3>
              <p>Join our Discord community for help and discussions</p>
              <a href="https://discord.gg/ollama" target="_blank" rel="noopener noreferrer">
                Join Discord
              </a>
            </div>
            <div className={styles.helpCard}>
              <h3>🐛 Report Issues</h3>
              <p>Found a bug or have a feature request?</p>
              <a href="https://github.com/ollama/ollama-distributed/issues" target="_blank" rel="noopener noreferrer">
                Report on GitHub
              </a>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}