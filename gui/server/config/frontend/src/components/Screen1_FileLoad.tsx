import React, { useState, useRef } from 'react';
import {
  Card,
  Button,
  Badge,
  Text,
  MessageBar,
  MessageBarBody,
  MessageBarTitle,
} from '@fluentui/react-components';
import {
  Folder24Regular,
  DocumentText24Regular,
  CheckmarkCircle24Filled,
  ArrowRight24Regular,
  Flash24Regular,
} from '@fluentui/react-icons';
import { AppConfig, DEFAULT_TOML_CONTENT } from '../types';
import { parseTOML } from '../tomlUtils';

interface Screen1Props {
  onConfigLoaded: (config: AppConfig, fileName: string, rawText: string) => void;
}

export const Screen1_FileLoad: React.FC<Screen1Props> = ({ onConfigLoaded }) => {
  const [isDragOver, setIsDragOver] = useState(false);
  const [loadedFileName, setLoadedFileName] = useState<string>('');
  const [loadedRawText, setLoadedRawText] = useState<string>('');
  const [parsedConfig, setParsedConfig] = useState<AppConfig | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const processFileContent = (content: string, name: string) => {
    try {
      const config = parseTOML(content);
      setParsedConfig(config);
      setLoadedFileName(name);
      setLoadedRawText(content);
      setErrorMsg(null);
    } catch (err: any) {
      setErrorMsg('Error al parsear el archivo TOML: ' + err.message);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const text = event.target?.result as string;
        processFileContent(text, file.name);
      };
      reader.readAsText(file);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragOver(false);
    const file = e.dataTransfer.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const text = event.target?.result as string;
        processFileContent(text, file.name);
      };
      reader.readAsText(file);
    }
  };

  const loadDefaultConfig = () => {
    processFileContent(DEFAULT_TOML_CONTENT, 'ageLANServer.toml');
  };

  const handleProceed = () => {
    if (parsedConfig && loadedRawText) {
      onConfigLoaded(parsedConfig, loadedFileName || 'ageLANServer.toml', loadedRawText);
    }
  };

  return (
    <div className="win-screen-container">
      {/* Header Banner */}
      <header className="win-header">
        <div className="win-header-badge">
          <DocumentText24Regular />
          <span>ageLANServer Configurator • Fluent UI v9</span>
        </div>
        <h1 className="win-title">Load Configuration File</h1>
        <p className="win-subtitle">
          Select or drag-and-drop the .toml configuration file to edit it with Fluent UI v9.
        </p>
      </header>

      {/* Dropzone Card */}
      <div
        className={`win-dropzone ${isDragOver ? 'drag-over' : ''} ${parsedConfig ? 'has-file' : ''}`}
        onDrop={handleDrop}
        onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
        onDragLeave={() => setIsDragOver(false)}
        onClick={() => fileInputRef.current?.click()}
      >
        <input
          type="file"
          ref={fileInputRef}
          onChange={handleFileChange}
          accept=".toml,.txt"
          style={{ display: 'none' }}
        />

        <div className="win-dropzone-content">
          <div className="win-dropzone-icon-wrapper">
            <Folder24Regular style={{ fontSize: 36 }} />
          </div>

          <div className="win-dropzone-text">
            <h3>Drag and drop your <span>.toml</span> file here</h3>
            <p>or click to browse your computer</p>
          </div>

          <div className="win-dropzone-actions" onClick={(e) => e.stopPropagation()}>
            <Button
              appearance="secondary"
              icon={<Folder24Regular />}
              onClick={() => fileInputRef.current?.click()}
            >
              Browse File
            </Button>
            <Button
              appearance="subtle"
              icon={<Flash24Regular />}
              onClick={loadDefaultConfig}
            >
              Load Default TOML
            </Button>
          </div>
        </div>
      </div>

      {errorMsg && (
        <MessageBar intent="error" style={{ marginBottom: 20 }}>
          <MessageBarBody>
            <MessageBarTitle>Error Loading TOML</MessageBarTitle>
            {errorMsg}
          </MessageBarBody>
        </MessageBar>
      )}

      {/* File Loaded Card Preview */}
      {parsedConfig && (
        <Card className="win-file-summary-card">
          <div className="win-file-summary-header">
            <div className="win-file-info">
              <div className="win-file-icon">
                <CheckmarkCircle24Filled />
              </div>
              <div>
                <h4 className="win-file-name">{loadedFileName}</h4>
                <Text size={200} style={{ color: 'rgba(255, 255, 255, 0.6)' }}>
                  TOML file ready for editing with Fluent UI v9
                </Text>
              </div>
            </div>
            <Badge appearance="tint" color="success" size="extra-large">
              Valid
            </Badge>
          </div>

          <div className="win-file-grid-details">
            <div className="win-detail-item">
              <span className="win-detail-label">Authentication</span>
              <span className="win-detail-val">{parsedConfig.Authentication}</span>
            </div>
            <div className="win-detail-item">
              <span className="win-detail-label">Enabled Games</span>
              <span className="win-detail-val">
                {parsedConfig.Games.Enabled.length ? parsedConfig.Games.Enabled.join(', ') : 'None (0)'}
              </span>
            </div>
            <div className="win-detail-item">
              <span className="win-detail-label">Announcement Port</span>
              <span className="win-detail-val">{parsedConfig.Announcement.Port}</span>
            </div>
            <div className="win-detail-item">
              <span className="win-detail-label">Multicast Group</span>
              <span className="win-detail-val">{parsedConfig.Announcement.MulticastGroup}</span>
            </div>
          </div>

          <div className="win-file-action-footer">
            <Button
              appearance="primary"
              size="large"
              icon={<ArrowRight24Regular />}
              iconPosition="after"
              onClick={handleProceed}
            >
              Open in Form Editor
            </Button>
          </div>
        </Card>
      )}
    </div>
  );
};
