import React from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { DataSourceHttpSettings } from '@grafana/ui';
import { ValkeyOptions } from '../types';

export type Props = DataSourcePluginOptionsEditorProps<ValkeyOptions>;

export const ValkeyConfigEditor = (props: Props) => {
    const { options, onOptionsChange } = props;

    return (
        <>
            <DataSourceHttpSettings
                defaultUrl="http://localhost:44322"
                dataSourceConfig={options}
                showAccessOptions={false}
                onChange={onOptionsChange}
            />
        </>
    );
};
