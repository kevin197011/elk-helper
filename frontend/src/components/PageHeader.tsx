// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Space, Typography, theme } from 'antd';

const { Title, Text } = Typography;

export interface PageHeaderProps {
  title: string;
  description?: string;
  extra?: React.ReactNode;
}

export default function PageHeader({ title, description, extra }: PageHeaderProps) {
  const { token } = theme.useToken();

  return (
    <div
      className="app-page-header"
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: description ? 'flex-start' : 'center',
        gap: 16,
        marginBottom: 16,
      }}
    >
      <Space direction="vertical" size={2} style={{ minWidth: 0 }}>
        <Title level={4} style={{ margin: 0 }}>
          {title}
        </Title>
        {description ? (
          <Text type="secondary" style={{ fontSize: 13, color: token.colorTextSecondary }}>
            {description}
          </Text>
        ) : null}
      </Space>

      {extra ? <div style={{ flexShrink: 0 }}>{extra}</div> : null}
    </div>
  );
}

