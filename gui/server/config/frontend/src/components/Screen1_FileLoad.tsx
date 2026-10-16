import React from 'react';
import {
    Button,
    MessageBar,
    MessageBarBody,
    MessageBarTitle,
} from '@fluentui/react-components';
import {
    Folder24Regular,
    DocumentText24Regular,
} from '@fluentui/react-icons';
import { ValidationError } from '../../bindings/changeme';

interface Screen1Props {
    onSelectConfig?: () => void;
    onCreateConfig?: () => void;
    errorMsg?: string;
    validationErrors?: ValidationError[];
}

export const Screen1_FileLoad: React.FC<Screen1Props> = ({ onSelectConfig, onCreateConfig, errorMsg, validationErrors = [] }) => {
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
                    Select the server's .toml configuration file to inspect its path and raw contents.
                </p>
            </header>

            {/* Dropzone Card */}
            <div
                id="config-dropzone"
                data-file-drop-target
                className="win-dropzone"
                onClick={() => onSelectConfig?.()}
            >
                <div className="win-dropzone-content">
                    <div className="win-dropzone-icon-wrapper">
                        <Folder24Regular style={{ fontSize: 36 }} />
                    </div>

                    <div className="win-dropzone-text">
                        <h3>Select your <span>config.toml</span> file</h3>
                         <p>or drag a TOML file here, or click to browse</p>
                    </div>

                    <div className="win-dropzone-actions" onClick={(e) => e.stopPropagation()}>
                        <Button
                            appearance="secondary"
                            icon={<Folder24Regular />}
                            onClick={() => onSelectConfig?.()}
                        >
                            Browse File
                        </Button>
                    </div>
                </div>
            </div>

            <div className="win-dropzone-actions" style={{ justifyContent: 'center', marginBottom: 20 }}>
                <Button appearance="primary" icon={<DocumentText24Regular />} onClick={() => onCreateConfig?.()}>
                    Create New Configuration
                </Button>
            </div>

            {errorMsg && (
                <MessageBar intent="error" style={{ marginBottom: 20 }}>
                    <MessageBarBody>
                        <MessageBarTitle>Error Loading TOML</MessageBarTitle>
                        {errorMsg}
                    </MessageBarBody>
                </MessageBar>
            )}

            {validationErrors.length > 0 && (
                <MessageBar intent="error" style={{ marginBottom: 20 }}>
                    <MessageBarBody>
                        <MessageBarTitle>Configuration validation errors</MessageBarTitle>
                        <ul>
                            {validationErrors.map((error) => (
                                <li key={`${error.field}-${error.message}`}>
                                    <strong>{error.field}</strong>: {error.message}
                                </li>
                            ))}
                        </ul>
                    </MessageBarBody>
                </MessageBar>
            )}

        </div>
    );
};
