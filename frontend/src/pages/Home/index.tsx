import {
  ApiOutlined,
  CloudServerOutlined,
  CodeOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  RocketOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Col, Row, Tag, Typography } from 'antd';
import React from 'react';
import styles from './index.less';

const { Text } = Typography;

const LAYERS = [
  {
    key: 'http',
    label: 'HTTP',
    desc: '接口层',
    detail: 'Gin + Swagger',
    color: '#22d3ee',
  },
  {
    key: 'app',
    label: 'App',
    desc: '应用层',
    detail: '用例编排',
    color: '#a78bfa',
  },
  {
    key: 'domain',
    label: 'Domain',
    desc: '领域层',
    detail: '核心业务',
    color: '#fb923c',
  },
  {
    key: 'infra',
    label: 'Infra',
    desc: '基础设施',
    detail: 'MySQL / Redis / Kafka',
    color: '#4ade80',
  },
];

const TECH = [
  { label: 'Go', color: '#00ADD8' },
  { label: 'Gin', color: '#00B386' },
  { label: 'GORM Gen', color: '#E44D26' },
  { label: 'Cobra', color: '#ED6A5A' },
  { label: 'MySQL', color: '#4479A1' },
  { label: 'Redis', color: '#DC382D' },
  { label: 'Kafka', color: '#231F20', text: '#fff' },
  { label: 'OpenTelemetry', color: '#F5A800' },
  { label: 'Ant Design', color: '#1677FF' },
  { label: 'UmiJS', color: '#1890FF' },
];

const CAPS = [
  {
    icon: <ApiOutlined />,
    title: 'RESTful API',
    desc: '版本化路由 · Swagger 文档自动生成',
  },
  {
    icon: <DeploymentUnitOutlined />,
    title: 'DDD 分层',
    desc: '严格单向依赖 · 清晰的领域边界',
  },
  {
    icon: <DatabaseOutlined />,
    title: '数据基础设施',
    desc: 'MySQL + Redis + Kafka 开箱即用',
  },
  {
    icon: <DashboardOutlined />,
    title: '可观测性',
    desc: 'OpenTelemetry 链路追踪 · Prometheus 指标',
  },
  {
    icon: <ThunderboltOutlined />,
    title: '代码生成',
    desc: 'GORM Gen 类型安全查询 · 零手写 DAO',
  },
  {
    icon: <CloudServerOutlined />,
    title: '容器部署',
    desc: 'Dockerfile + Docker Compose 一键编排',
  },
  {
    icon: <CodeOutlined />,
    title: 'AI 协作',
    desc: 'Knowledge 文档驱动 · AI 精准理解架构',
  },
  {
    icon: <RocketOutlined />,
    title: '工程化',
    desc: 'Makefile · golangci-lint · 健康检查探针',
  },
];

const Home: React.FC = () => {
  return (
    <div className={styles.page}>
      {/* Hero */}
      <div className={styles.hero}>
        <div className={styles.gridOverlay} />
        <div className={styles.heroContent}>
          <div className={styles.brand}>
            <span className={styles.brandQ}>Q</span>
            <span className={styles.brandDash}>-</span>
            <span className={styles.brandDev}>DEV</span>
          </div>
          <div className={styles.tagline}>AI 驱动的全栈开发脚手架</div>
          <div className={styles.subtitle}>
            让 AI 理解你的架构，让开发回归本质
          </div>
          <div className={styles.version}>
            <Tag
              bordered={false}
              style={{
                background: 'rgba(34,211,238,0.15)',
                color: '#22d3ee',
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 12,
              }}
            >
              v0.1.0
            </Tag>
            <Text
              style={{
                color: 'rgba(255,255,255,0.35)',
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 12,
              }}
            >
              Go · Gin · GORM · Ant Design Pro
            </Text>
          </div>
        </div>
      </div>

      {/* Architecture Flow */}
      <div className={styles.section}>
        <div className={styles.sectionHeader}>
          <span className={styles.sectionTag}>ARCHITECTURE</span>
          <span className={styles.sectionTitle}>DDD 分层架构</span>
        </div>
        <div className={styles.archFlow}>
          {LAYERS.map((layer, i) => (
            <React.Fragment key={layer.key}>
              <div className={styles.archNode}>
                <div
                  className={styles.archNodeBar}
                  style={{ background: layer.color }}
                />
                <div className={styles.archNodeBody}>
                  <div
                    className={styles.archNodeLabel}
                    style={{ color: layer.color }}
                  >
                    {layer.label}
                  </div>
                  <div className={styles.archNodeDesc}>{layer.desc}</div>
                  <div className={styles.archNodeDetail}>{layer.detail}</div>
                </div>
              </div>
              {i < LAYERS.length - 1 && (
                <div className={styles.archArrow}>
                  <svg width="32" height="16" viewBox="0 0 32 16">
                    <path
                      d="M0 8 L24 8 M20 3 L26 8 L20 13"
                      stroke="rgba(255,255,255,0.2)"
                      strokeWidth="1.5"
                      fill="none"
                    />
                  </svg>
                </div>
              )}
            </React.Fragment>
          ))}
        </div>
        <div className={styles.archNote}>
          <CodeOutlined style={{ marginRight: 6, opacity: 0.5 }} />
          严格单向依赖：http → app → domain → infra
        </div>
      </div>

      {/* Tech Stack */}
      <div className={styles.section}>
        <div className={styles.sectionHeader}>
          <span className={styles.sectionTag}>STACK</span>
          <span className={styles.sectionTitle}>技术栈</span>
        </div>
        <div className={styles.techGrid}>
          {TECH.map((t) => (
            <div key={t.label} className={styles.techBadge}>
              <span
                className={styles.techDot}
                style={{ background: t.color }}
              />
              <span className={styles.techLabel}>{t.label}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Capabilities */}
      <div className={styles.section}>
        <div className={styles.sectionHeader}>
          <span className={styles.sectionTag}>CAPABILITIES</span>
          <span className={styles.sectionTitle}>核心能力</span>
        </div>
        <Row gutter={[16, 16]}>
          {CAPS.map((cap) => (
            <Col key={cap.title} xs={24} sm={12} lg={6}>
              <div className={styles.capCard}>
                <div className={styles.capIcon}>{cap.icon}</div>
                <div className={styles.capTitle}>{cap.title}</div>
                <div className={styles.capDesc}>{cap.desc}</div>
              </div>
            </Col>
          ))}
        </Row>
      </div>
    </div>
  );
};

export default Home;
