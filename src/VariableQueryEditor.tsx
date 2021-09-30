import { DataSource } from 'datasource';
import React from 'react';
import { QueryEditor } from './QueryEditor';
import { MyQuery } from './types';

interface Props {
  query: MyQuery;
  onChange: (query: MyQuery, definition?: string) => void;
  datasource: DataSource;
}

export const VariableQueryEditor = ({ onChange, query, datasource }: Props) => {
  const saveQuery = (newQuery: MyQuery) => {
    if (newQuery) {
      onChange(newQuery, newQuery.queryText);
    }
  };

  return <QueryEditor onRunQuery={() => {}} onChange={saveQuery} query={{ ...query }} datasource={datasource} />;
};
