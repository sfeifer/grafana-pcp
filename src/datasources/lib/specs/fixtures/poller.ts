import { RecursivePartial } from 'common/types/utils';
import { EndpointState, EndpointWithCtx } from 'datasources/lib/pmapi/poller/types';
import { Target, TargetState } from 'datasources/lib/pmapi/types';
import { defaultsDeep } from 'lodash';
import { datasource } from '.';

export function endpoint(props?: RecursivePartial<EndpointWithCtx>): EndpointWithCtx {
    return defaultsDeep(props, {
        state: EndpointState.CONNECTED,
        url: '',
        hostspec: '',
        metrics: [],
        targets: [],
        additionalMetricsToPoll: [],
        errors: [],
        context: {
            context: 123,
            labels: {},
        },
    });
}

export function target(props?: RecursivePartial<Target>): Target {
    const query = datasource.templatedQuery(props?.query);
    return defaultsDeep(props, {
        targetId: `0/1/${query.refId ?? 'A'}`,
        state: TargetState.METRICS_AVAILABLE,
        query,
        metricNames: [query.expr],
        errors: [],
        lastActiveMs: 0,
    });
}
