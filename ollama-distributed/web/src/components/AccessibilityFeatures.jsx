import React, { useState, useEffect, useCallback } from 'react';
import { Card, Button, Form, Modal, Alert, Badge, ButtonGroup } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faUniversalAccess,
  faEye,
  faEyeSlash,
  faVolumeUp,
  faVolumeOff,
  faKeyboard,
  faMousePointer,
  faPalette,
  faTextHeight,
  faAdjust,
  faExpand,
  faCompress,
  faCog,
  faPlay,
  faPause,
  faQuestionCircle,
  faCheck,
  faExclamationTriangle,
  faMicrophone,
  faMicrophoneSlash,
  faClosedCaptioning,
  faLanguage
} from '@fortawesome/free-solid-svg-icons';

const AccessibilityFeatures = ({
  onSettingsChange,
  initialSettings = {},
  className = ""
}) => {
  const [settings, setSettings] = useState({
    highContrast: false,
    largeText: false,
    reducedMotion: false,
    screenReader: false,
    voiceNavigation: false,
    keyboardNavigation: true,
    focusIndicator: true,
    colorBlindSupport: false,
    customColors: {
      primary: '#3B82F6',
      secondary: '#10B981',
      background: '#FFFFFF',
      text: '#111827'
    },
    fontSize: 16,
    lineHeight: 1.5,
    letterSpacing: 0,
    audioDescriptions: false,
    captions: false,
    signLanguage: false,
    language: 'en',
    ...initialSettings
  });
  
  const [showSettings, setShowSettings] = useState(false);
  const [showHelp, setShowHelp] = useState(false);
  const [isListening, setIsListening] = useState(false);
  const [speechSynthesis, setSpeechSynthesis] = useState(null);
  const [speechRecognition, setSpeechRecognition] = useState(null);
  const [announcement, setAnnouncement] = useState('');
  const [keyboardShortcuts, setKeyboardShortcuts] = useState({
    'Alt+H': 'Open help',
    'Alt+S': 'Open settings',
    'Alt+C': 'Toggle high contrast',
    'Alt+T': 'Toggle large text',
    'Alt+M': 'Toggle reduced motion',
    'Alt+R': 'Toggle screen reader',
    'Alt+V': 'Toggle voice navigation',
    'Escape': 'Close modal/menu',
    'Enter/Space': 'Activate button',
    'Tab': 'Navigate forward',
    'Shift+Tab': 'Navigate backward'
  });

  // Initialize accessibility features
  useEffect(() => {
    initializeAccessibilityAPIs();
    applyAccessibilitySettings(settings);
    setupKeyboardShortcuts();
    
    return () => {
      cleanupAccessibilityFeatures();
    };
  }, []);

  // Apply settings when they change
  useEffect(() => {
    applyAccessibilitySettings(settings);
    if (onSettingsChange) {
      onSettingsChange(settings);
    }
  }, [settings, onSettingsChange]);

  const initializeAccessibilityAPIs = () => {
    // Initialize Speech Synthesis
    if ('speechSynthesis' in window) {
      setSpeechSynthesis(window.speechSynthesis);
    }

    // Initialize Speech Recognition
    if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
      const recognition = new SpeechRecognition();
      recognition.continuous = true;
      recognition.interimResults = true;
      recognition.lang = settings.language;
      
      recognition.onstart = () => setIsListening(true);
      recognition.onend = () => setIsListening(false);
      recognition.onerror = (event) => {
        console.error('Speech recognition error:', event.error);
        setIsListening(false);
      };
      recognition.onresult = handleSpeechResult;
      
      setSpeechRecognition(recognition);
    }
  };

  const handleSpeechResult = (event) => {
    const lastResult = event.results[event.results.length - 1];
    if (lastResult.isFinal) {
      const command = lastResult[0].transcript.toLowerCase().trim();
      processSpeechCommand(command);
    }
  };

  const processSpeechCommand = (command) => {
    const commands = {
      'help': () => setShowHelp(true),
      'settings': () => setShowSettings(true),
      'close': () => { setShowSettings(false); setShowHelp(false); },
      'high contrast': () => toggleSetting('highContrast'),
      'large text': () => toggleSetting('largeText'),
      'reduce motion': () => toggleSetting('reducedMotion'),
      'screen reader': () => toggleSetting('screenReader'),
      'read page': () => readPageContent(),
      'stop reading': () => stopSpeaking()
    };

    const matchedCommand = Object.keys(commands).find(cmd => 
      command.includes(cmd)
    );

    if (matchedCommand) {
      commands[matchedCommand]();
      announce(`Executed: ${matchedCommand}`);
    } else {
      announce('Command not recognized. Say "help" for available commands.');
    }
  };

  const applyAccessibilitySettings = (newSettings) => {
    const root = document.documentElement;
    
    // High contrast mode
    root.classList.toggle('high-contrast', newSettings.highContrast);
    
    // Large text
    root.classList.toggle('large-text', newSettings.largeText);
    root.style.setProperty('--base-font-size', `${newSettings.fontSize}px`);
    root.style.setProperty('--line-height', newSettings.lineHeight);
    root.style.setProperty('--letter-spacing', `${newSettings.letterSpacing}px`);
    
    // Reduced motion
    root.classList.toggle('reduced-motion', newSettings.reducedMotion);
    
    // Focus indicator
    root.classList.toggle('enhanced-focus', newSettings.focusIndicator);
    
    // Color blind support
    root.classList.toggle('color-blind-support', newSettings.colorBlindSupport);
    
    // Custom colors
    if (newSettings.customColors) {
      Object.entries(newSettings.customColors).forEach(([key, value]) => {
        root.style.setProperty(`--accessibility-${key}`, value);
      });
    }
    
    // Keyboard navigation
    root.classList.toggle('keyboard-navigation', newSettings.keyboardNavigation);
  };

  const setupKeyboardShortcuts = () => {
    const handleKeyDown = (event) => {
      const key = event.key;
      const modifiers = [];
      
      if (event.altKey) modifiers.push('Alt');
      if (event.ctrlKey) modifiers.push('Ctrl');
      if (event.shiftKey) modifiers.push('Shift');
      if (event.metaKey) modifiers.push('Meta');
      
      const shortcut = modifiers.length > 0 
        ? `${modifiers.join('+')}+${key}` 
        : key;

      switch (shortcut) {
        case 'Alt+h':
        case 'Alt+H':
          event.preventDefault();
          setShowHelp(true);
          break;
        case 'Alt+s':
        case 'Alt+S':
          event.preventDefault();
          setShowSettings(true);
          break;
        case 'Alt+c':
        case 'Alt+C':
          event.preventDefault();
          toggleSetting('highContrast');
          break;
        case 'Alt+t':
        case 'Alt+T':
          event.preventDefault();
          toggleSetting('largeText');
          break;
        case 'Alt+m':
        case 'Alt+M':
          event.preventDefault();
          toggleSetting('reducedMotion');
          break;
        case 'Alt+r':
        case 'Alt+R':
          event.preventDefault();
          toggleSetting('screenReader');
          break;
        case 'Alt+v':
        case 'Alt+V':
          event.preventDefault();
          toggleVoiceNavigation();
          break;
        case 'Escape':
          if (showSettings) setShowSettings(false);
          if (showHelp) setShowHelp(false);
          break;
        default:
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  };

  const cleanupAccessibilityFeatures = () => {
    if (speechSynthesis) {
      speechSynthesis.cancel();
    }
    if (speechRecognition && isListening) {
      speechRecognition.stop();
    }
  };

  const toggleSetting = (settingName) => {
    setSettings(prev => ({
      ...prev,
      [settingName]: !prev[settingName]
    }));
    announce(`${settingName} ${!settings[settingName] ? 'enabled' : 'disabled'}`);
  };

  const updateSetting = (settingName, value) => {
    setSettings(prev => ({
      ...prev,
      [settingName]: value
    }));
  };

  const toggleVoiceNavigation = () => {
    if (!speechRecognition) {
      announce('Speech recognition is not supported in this browser');
      return;
    }

    if (isListening) {
      speechRecognition.stop();
      updateSetting('voiceNavigation', false);
    } else {
      speechRecognition.start();
      updateSetting('voiceNavigation', true);
    }
  };

  const announce = (message) => {
    setAnnouncement(message);
    
    if (settings.screenReader && speechSynthesis) {
      const utterance = new SpeechSynthesisUtterance(message);
      utterance.rate = 0.8;
      utterance.volume = 0.8;
      speechSynthesis.speak(utterance);
    }

    // Clear announcement after 3 seconds
    setTimeout(() => setAnnouncement(''), 3000);
  };

  const readPageContent = () => {
    if (!speechSynthesis) {
      announce('Text-to-speech is not supported in this browser');
      return;
    }

    const textContent = document.body.innerText;
    const utterance = new SpeechSynthesisUtterance(textContent);
    utterance.rate = 0.8;
    utterance.volume = 0.8;
    speechSynthesis.speak(utterance);
  };

  const stopSpeaking = () => {
    if (speechSynthesis) {
      speechSynthesis.cancel();
    }
  };

  const runAccessibilityAudit = () => {
    const issues = [];
    
    // Check for missing alt text
    const images = document.querySelectorAll('img:not([alt])');
    if (images.length > 0) {
      issues.push(`${images.length} images missing alt text`);
    }
    
    // Check for proper heading structure
    const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
    let prevLevel = 0;
    headings.forEach(heading => {
      const level = parseInt(heading.tagName[1]);
      if (level > prevLevel + 1) {
        issues.push(`Heading level skipped: ${heading.tagName}`);
      }
      prevLevel = level;
    });
    
    // Check for keyboard accessible buttons
    const buttons = document.querySelectorAll('button, [role="button"]');
    buttons.forEach(button => {
      if (!button.hasAttribute('tabindex') && button.tabIndex < 0) {
        issues.push('Button not keyboard accessible');
      }
    });
    
    return issues;
  };

  const AccessibilityToolbar = () => (
    <div className="accessibility-toolbar position-fixed top-0 end-0 p-3" style={{ zIndex: 1060 }}>
      <ButtonGroup vertical size="sm">
        <Button
          variant={settings.highContrast ? 'success' : 'outline-secondary'}
          onClick={() => toggleSetting('highContrast')}
          title="High Contrast (Alt+C)"
        >
          <FontAwesomeIcon icon={faAdjust} />
        </Button>
        
        <Button
          variant={settings.largeText ? 'success' : 'outline-secondary'}
          onClick={() => toggleSetting('largeText')}
          title="Large Text (Alt+T)"
        >
          <FontAwesomeIcon icon={faTextHeight} />
        </Button>
        
        <Button
          variant={settings.reducedMotion ? 'success' : 'outline-secondary'}
          onClick={() => toggleSetting('reducedMotion')}
          title="Reduce Motion (Alt+M)"
        >
          <FontAwesomeIcon icon={faCompress} />
        </Button>
        
        <Button
          variant={settings.screenReader ? 'success' : 'outline-secondary'}
          onClick={() => toggleSetting('screenReader')}
          title="Screen Reader (Alt+R)"
        >
          <FontAwesomeIcon icon={settings.screenReader ? faVolumeUp : faVolumeOff} />
        </Button>
        
        <Button
          variant={isListening ? 'success' : 'outline-secondary'}
          onClick={toggleVoiceNavigation}
          title="Voice Navigation (Alt+V)"
          disabled={!speechRecognition}
        >
          <FontAwesomeIcon icon={isListening ? faMicrophone : faMicrophoneSlash} />
        </Button>
        
        <Button
          variant="outline-primary"
          onClick={() => setShowSettings(true)}
          title="Settings (Alt+S)"
        >
          <FontAwesomeIcon icon={faCog} />
        </Button>
        
        <Button
          variant="outline-info"
          onClick={() => setShowHelp(true)}
          title="Help (Alt+H)"
        >
          <FontAwesomeIcon icon={faQuestionCircle} />
        </Button>
      </ButtonGroup>
    </div>
  );

  const SettingsModal = () => (
    <Modal show={showSettings} onHide={() => setShowSettings(false)} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>
          <FontAwesomeIcon icon={faUniversalAccess} className="me-2" />
          Accessibility Settings
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="row">
          <div className="col-md-6">
            <h6>Visual</h6>
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="High Contrast Mode"
                checked={settings.highContrast}
                onChange={(e) => updateSetting('highContrast', e.target.checked)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Large Text"
                checked={settings.largeText}
                onChange={(e) => updateSetting('largeText', e.target.checked)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Font Size: {settings.fontSize}px</Form.Label>
              <Form.Range
                min={12}
                max={24}
                value={settings.fontSize}
                onChange={(e) => updateSetting('fontSize', parseInt(e.target.value))}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Line Height: {settings.lineHeight}</Form.Label>
              <Form.Range
                min={1.2}
                max={2.0}
                step={0.1}
                value={settings.lineHeight}
                onChange={(e) => updateSetting('lineHeight', parseFloat(e.target.value))}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Color Blind Support"
                checked={settings.colorBlindSupport}
                onChange={(e) => updateSetting('colorBlindSupport', e.target.checked)}
              />
            </Form.Group>
          </div>
          
          <div className="col-md-6">
            <h6>Motion & Interaction</h6>
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Reduce Motion"
                checked={settings.reducedMotion}
                onChange={(e) => updateSetting('reducedMotion', e.target.checked)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Enhanced Focus Indicator"
                checked={settings.focusIndicator}
                onChange={(e) => updateSetting('focusIndicator', e.target.checked)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Keyboard Navigation"
                checked={settings.keyboardNavigation}
                onChange={(e) => updateSetting('keyboardNavigation', e.target.checked)}
              />
            </Form.Group>
            
            <h6 className="mt-4">Audio</h6>
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Screen Reader Support"
                checked={settings.screenReader}
                onChange={(e) => updateSetting('screenReader', e.target.checked)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Voice Navigation"
                checked={settings.voiceNavigation}
                onChange={(e) => updateSetting('voiceNavigation', e.target.checked)}
                disabled={!speechRecognition}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Check
                type="switch"
                label="Audio Descriptions"
                checked={settings.audioDescriptions}
                onChange={(e) => updateSetting('audioDescriptions', e.target.checked)}
              />
            </Form.Group>
          </div>
        </div>
        
        <div className="mt-4">
          <h6>Accessibility Audit</h6>
          <Button variant="outline-info" onClick={() => {
            const issues = runAccessibilityAudit();
            if (issues.length === 0) {
              announce('No accessibility issues found');
            } else {
              announce(`Found ${issues.length} accessibility issues`);
            }
          }}>
            <FontAwesomeIcon icon={faCheck} className="me-2" />
            Run Accessibility Audit
          </Button>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={() => setShowSettings(false)}>
          Close
        </Button>
        <Button variant="primary" onClick={() => {
          localStorage.setItem('accessibilitySettings', JSON.stringify(settings));
          announce('Settings saved');
        }}>
          Save Settings
        </Button>
      </Modal.Footer>
    </Modal>
  );

  const HelpModal = () => (
    <Modal show={showHelp} onHide={() => setShowHelp(false)} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>
          <FontAwesomeIcon icon={faQuestionCircle} className="me-2" />
          Accessibility Help
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="row">
          <div className="col-md-6">
            <h6>Keyboard Shortcuts</h6>
            <div className="table-responsive">
              <table className="table table-sm">
                <tbody>
                  {Object.entries(keyboardShortcuts).map(([shortcut, description]) => (
                    <tr key={shortcut}>
                      <td><Badge bg="secondary">{shortcut}</Badge></td>
                      <td>{description}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          
          <div className="col-md-6">
            <h6>Voice Commands</h6>
            <ul className="list-unstyled">
              <li><Badge bg="info">"Help"</Badge> - Open this help dialog</li>
              <li><Badge bg="info">"Settings"</Badge> - Open accessibility settings</li>
              <li><Badge bg="info">"High contrast"</Badge> - Toggle high contrast mode</li>
              <li><Badge bg="info">"Large text"</Badge> - Toggle large text mode</li>
              <li><Badge bg="info">"Reduce motion"</Badge> - Toggle motion reduction</li>
              <li><Badge bg="info">"Screen reader"</Badge> - Toggle screen reader</li>
              <li><Badge bg="info">"Read page"</Badge> - Read page content aloud</li>
              <li><Badge bg="info">"Stop reading"</Badge> - Stop text-to-speech</li>
              <li><Badge bg="info">"Close"</Badge> - Close open dialogs</li>
            </ul>
            
            <h6 className="mt-4">Features</h6>
            <ul>
              <li><strong>High Contrast:</strong> Increases color contrast for better visibility</li>
              <li><strong>Large Text:</strong> Increases font size for easier reading</li>
              <li><strong>Reduced Motion:</strong> Minimizes animations and transitions</li>
              <li><strong>Screen Reader:</strong> Provides audio feedback for actions</li>
              <li><strong>Voice Navigation:</strong> Control the interface with voice commands</li>
              <li><strong>Keyboard Navigation:</strong> Full keyboard accessibility support</li>
            </ul>
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={() => setShowHelp(false)}>
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  );

  return (
    <div className={`accessibility-features ${className}`}>
      {/* Live region for announcements */}
      <div 
        className="sr-only" 
        aria-live="polite" 
        aria-atomic="true"
        role="status"
      >
        {announcement}
      </div>
      
      {/* Skip navigation link */}
      <a 
        href="#main-content" 
        className="skip-link position-absolute"
        style={{
          top: '-40px',
          left: '6px',
          zIndex: 2000,
          padding: '8px',
          backgroundColor: '#000',
          color: '#fff',
          textDecoration: 'none',
          borderRadius: '0 0 4px 4px'
        }}
        onFocus={(e) => {
          e.target.style.top = '6px';
        }}
        onBlur={(e) => {
          e.target.style.top = '-40px';
        }}
      >
        Skip to main content
      </a>
      
      {/* Accessibility Toolbar */}
      <AccessibilityToolbar />
      
      {/* Settings Modal */}
      <SettingsModal />
      
      {/* Help Modal */}
      <HelpModal />
      
      {/* Voice Navigation Status */}
      {isListening && (
        <div 
          className="voice-status position-fixed bottom-0 end-0 m-3 p-2 bg-success text-white rounded"
          style={{ zIndex: 1050 }}
        >
          <FontAwesomeIcon icon={faMicrophone} className="me-2" />
          Listening... Speak a command
        </div>
      )}
    </div>
  );
};

export default AccessibilityFeatures;