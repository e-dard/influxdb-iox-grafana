import { DataSourceInstanceSettings, DataFrame, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }
  applyTemplateVariables(query: MyQuery) {
    const templateSrv = getTemplateSrv();
    return {
      ...query,
      queryText: query.queryText ? templateSrv.replace(query.queryText) : '',
    };
  }

  // this is basically taken verbatim from the Grafana CSV plugin example.
  async metricFindQuery(query: MyQuery, options?: any) {
    const request = {
      targets: [
        {
          ...query,
          refId: 'metricFindQuery',
        },
      ],
      range: options.range,
      rangeRaw: options.rangeRaw,
    } as DataQueryRequest<MyQuery>;

    let res: DataQueryResponse | undefined;

    try {
      res = await this.query(request).toPromise();
    } catch (err) {
      return Promise.reject(err);
    }

    res = res!;
    if (!res.data.length || !res.data[0].fields.length) {
      return [];
    }

    return (res.data[0] as DataFrame).fields[0].values.toArray().map((_) => ({ text: _.toString() }));
  }
}
