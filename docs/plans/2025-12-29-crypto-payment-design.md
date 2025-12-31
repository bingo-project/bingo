# 加密货币支付设计

## 概述

为 Bingo 项目添加加密货币支付功能，支持用户使用已登录的钱包进行一次性支付和手动续费。

## 设计决策

| 决策点 | 选择 | 理由 |
|-------|------|------|
| 支付场景 | 一次性支付 + 手动续费 | 订阅自动扣款走信用卡，加密货币不支持自动扣款 |
| 技术方案 | 简单转账 + TxHash 提交 | 无需智能合约，Gas 费低，实现简单 |
| 钱包限制 | 必须用登录钱包支付 | 利用 SIWE 已知用户地址，简化订单关联 |
| 地址策略 | 固定收款地址 | 通过 from 地址识别用户，无需管理多地址 |
| 确认机制 | 前端提交 TxHash + 后端兜底扫链 | 兼顾体验和可靠性 |

## 整体流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                           支付流程                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │  创建    │    │  发起    │    │  提交    │    │  验证    │      │
│  │  订单    │ -> │  转账    │ -> │  TxHash  │ -> │  完成    │      │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘      │
│       │               │               │               │             │
│       v               v               v               v             │
│   后端生成        前端调起         前端提交        后端查链          │
│   订单号+金额     MetaMask        交易哈希        验证到账          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 时序图

```
用户          前端                后端              区块链
 │             │                   │                  │
 │─创建订单───>│                   │                  │
 │             │──POST /orders────>│                  │
 │             │<─{orderId,amount}─│                  │
 │             │                   │                  │
 │<─弹出钱包───│                   │                  │
 │─确认支付───>│                   │                  │
 │             │───────────────sendTransaction──────->│
 │             │<──────────────────txHash─────────────│
 │             │                   │                  │
 │             │─POST /orders/:id/confirm {txHash}──>│
 │             │                   │────查询交易─────>│
 │             │                   │<───交易详情──────│
 │             │<──────{success}───│                  │
 │<─支付成功───│                   │                  │
```

## 后端 API 设计

### API 端点

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | `/v1/orders` | 创建订单 |
| GET | `/v1/orders/:id` | 查询订单状态 |
| POST | `/v1/orders/:id/confirm` | 提交 TxHash 确认支付 |

### 创建订单

**POST /v1/orders**

```json
// Request
{
  "productId": "pro_monthly",
  "chainId": 1,
  "currency": "ETH"
}

// Response
{
  "orderId": "ord_20251229_abc123",
  "amount": "0.005",
  "currency": "ETH",
  "chainId": 1,
  "recipient": "0x1234...abcd",
  "expiresAt": "2025-12-29T11:00:00Z",
  "status": "pending"
}
```

### 确认支付

**POST /v1/orders/:id/confirm**

```json
// Request
{
  "txHash": "0xabc123..."
}

// Response
{
  "orderId": "ord_20251229_abc123",
  "status": "paid",
  "confirmedAt": "2025-12-29T10:35:00Z"
}
```

### 后端验证逻辑

```go
func (b *orderBiz) Confirm(ctx context.Context, orderID string, req *v1.ConfirmRequest) error {
    // 1. 获取订单
    order, err := b.orderStore.Get(ctx, orderID)
    if order.Status != model.OrderStatusPending {
        return errno.ErrOrderNotPending
    }

    // 2. 检查 txHash 是否已被使用（防重放）
    if b.orderStore.TxHashExists(ctx, req.TxHash) {
        return errno.ErrTxHashAlreadyUsed
    }

    // 3. 查询链上交易
    tx, err := b.chainClient.GetTransaction(ctx, order.ChainID, req.TxHash)
    if err != nil {
        return errno.ErrTxNotFound
    }

    // 4. 验证交易状态
    if tx.Status != 1 {
        return errno.ErrTxFailed
    }

    // 5. 验证收款地址
    if !strings.EqualFold(tx.To, b.cfg.PaymentAddresses[order.ChainID]) {
        return errno.ErrInvalidRecipient
    }

    // 6. 验证发送方 = 用户登录钱包
    user, _ := b.userStore.Get(ctx, order.UID)
    if !strings.EqualFold(tx.From, user.WalletAddress) {
        return errno.ErrInvalidSender
    }

    // 7. 验证金额（允许 1% 误差，应对汇率波动）
    if tx.Value.Cmp(order.Amount.Mul(0.99)) < 0 {
        return errno.ErrInsufficientAmount
    }

    // 8. 验证确认数
    if tx.Confirmations < b.cfg.RequiredConfirmations {
        return errno.ErrInsufficientConfirmations
    }

    // 9. 更新订单状态
    return b.orderStore.MarkPaid(ctx, orderID, req.TxHash)
}
```

### 配置结构

```yaml
# bingo-apiserver.yaml
payment:
  addresses:
    1: "0x1234...abcd"      # Ethereum Mainnet
    56: "0x1234...abcd"     # BSC
    137: "0x1234...abcd"    # Polygon

  requiredConfirmations:
    1: 12      # Ethereum ~3分钟
    56: 15     # BSC ~45秒
    137: 128   # Polygon ~4分钟

  orderExpiration: 30m

  supportedCurrencies:
    - ETH
    - USDT
    - USDC
```

## 数据结构

### 订单表 `uc_order`

```sql
CREATE TABLE `uc_order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` varchar(32) NOT NULL COMMENT '订单号',
  `uid` varchar(64) NOT NULL COMMENT '用户ID',
  `product_id` varchar(64) NOT NULL COMMENT '商品ID',
  `amount` decimal(36,18) NOT NULL COMMENT '支付金额（加密货币）',
  `currency` varchar(16) NOT NULL COMMENT '币种 ETH/USDT/USDC',
  `chain_id` int NOT NULL COMMENT '链ID',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/paid/expired/refunded',
  `tx_hash` varchar(66) DEFAULT NULL COMMENT '交易哈希',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `expires_at` datetime NOT NULL COMMENT '订单过期时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_id` (`order_id`),
  UNIQUE KEY `uk_tx_hash` (`tx_hash`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status_expires` (`status`, `expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
```

### 商品表 `uc_product`

```sql
CREATE TABLE `uc_product` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `product_id` varchar(64) NOT NULL COMMENT '商品ID',
  `name` varchar(128) NOT NULL COMMENT '商品名称',
  `description` text COMMENT '商品描述',
  `price_usd` decimal(10,2) NOT NULL COMMENT '美元价格',
  `type` varchar(16) NOT NULL COMMENT 'one_time/subscription',
  `status` varchar(16) NOT NULL DEFAULT 'active' COMMENT 'active/inactive',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';
```

### 订单状态流转

```
                 ┌─────────────┐
                 │   pending   │
                 └──────┬──────┘
                        │
           ┌────────────┼────────────┐
           │            │            │
           v            v            v
     ┌──────────┐ ┌──────────┐ ┌──────────┐
     │   paid   │ │  expired │ │ cancelled│
     └──────────┘ └──────────┘ └──────────┘
           │
           v
     ┌──────────┐
     │ refunded │  (未来可能需要)
     └──────────┘
```

### Model 定义

```go
// internal/apiserver/model/order.go

type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusPaid      OrderStatus = "paid"
    OrderStatusExpired   OrderStatus = "expired"
    OrderStatusCancelled OrderStatus = "cancelled"
    OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
    ID        uint64          `gorm:"primaryKey"`
    OrderID   string          `gorm:"uniqueIndex;size:32"`
    UID       string          `gorm:"index;size:64"`
    ProductID string          `gorm:"size:64"`
    Amount    decimal.Decimal `gorm:"type:decimal(36,18)"`
    Currency  string          `gorm:"size:16"`
    ChainID   int
    Status    OrderStatus     `gorm:"size:16;index"`
    TxHash    *string         `gorm:"uniqueIndex;size:66"`
    PaidAt    *time.Time
    ExpiresAt time.Time       `gorm:"index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## 前端实现

### 收款地址白名单（硬编码）

```typescript
// config/payment.ts

export const PAYMENT_ADDRESSES: Record<number, string> = {
  1: "0x1234567890abcdef1234567890abcdef12345678",      // Ethereum
  56: "0x1234567890abcdef1234567890abcdef12345678",     // BSC
  137: "0x1234567890abcdef1234567890abcdef12345678",    // Polygon
};

export const SUPPORTED_CHAINS = [
  { chainId: 1, name: "Ethereum", currency: "ETH", icon: "/icons/eth.svg" },
  { chainId: 56, name: "BNB Chain", currency: "BNB", icon: "/icons/bnb.svg" },
  { chainId: 137, name: "Polygon", currency: "MATIC", icon: "/icons/matic.svg" },
];
```

### 支付流程

```typescript
// hooks/usePayment.ts

import { useSendTransaction, useWaitForTransaction } from 'wagmi';
import { parseEther } from 'viem';
import { PAYMENT_ADDRESSES } from '@/config/payment';

export function usePayment() {
  const { sendTransactionAsync } = useSendTransaction();

  async function pay(productId: string, chainId: number) {
    // 1. 创建订单
    const order = await api.post('/v1/orders', {
      productId,
      chainId,
      currency: 'ETH',
    });

    // 2. 验证后端返回地址与硬编码一致
    const expectedAddress = PAYMENT_ADDRESSES[chainId];
    if (order.recipient.toLowerCase() !== expectedAddress.toLowerCase()) {
      throw new Error('收款地址异常，请联系客服');
    }

    // 3. 发起链上交易
    const txHash = await sendTransactionAsync({
      to: expectedAddress,
      value: parseEther(order.amount),
      chainId,
    });

    // 4. 等待交易确认（至少 1 个区块）
    await waitForTransaction({ hash: txHash });

    // 5. 提交 txHash 到后端
    const result = await api.post(`/v1/orders/${order.orderId}/confirm`, {
      txHash,
    });

    return result;
  }

  return { pay };
}
```

### 支付组件

```tsx
// components/PaymentButton.tsx

export function PaymentButton({ productId }: { productId: string }) {
  const { pay } = usePayment();
  const { chain } = useNetwork();
  const [status, setStatus] = useState<'idle' | 'paying' | 'confirming' | 'success' | 'error'>('idle');

  const handlePay = async () => {
    try {
      setStatus('paying');
      await pay(productId, chain.id);
      setStatus('success');
    } catch (e) {
      setStatus('error');
    }
  };

  return (
    <div>
      <button onClick={handlePay} disabled={status === 'paying' || status === 'confirming'}>
        {status === 'idle' && '使用加密货币支付'}
        {status === 'paying' && '请在钱包中确认...'}
        {status === 'confirming' && '交易确认中...'}
        {status === 'success' && '支付成功 ✓'}
        {status === 'error' && '支付失败，点击重试'}
      </button>

      {/* 安全提示 */}
      <p className="text-sm text-gray-500 mt-2">
        请在钱包中确认收款地址为：
        <code className="bg-gray-100 px-1">{PAYMENT_ADDRESSES[chain.id]?.slice(0, 10)}...</code>
      </p>
    </div>
  );
}
```

### 用户关闭页面兜底

```typescript
// hooks/usePayment.ts (补充)

async function pay(productId: string, chainId: number) {
  // ... 前面的代码

  // 3. 发起链上交易
  const txHash = await sendTransactionAsync({...});

  // 3.5 立即保存 txHash 到 localStorage（防止用户关页面）
  localStorage.setItem(`pending_tx_${order.orderId}`, txHash);

  // 4. 等待交易确认
  await waitForTransaction({ hash: txHash });

  // 5. 提交确认
  const result = await api.post(`/v1/orders/${order.orderId}/confirm`, { txHash });

  // 6. 清除 localStorage
  localStorage.removeItem(`pending_tx_${order.orderId}`);

  return result;
}

// 页面加载时检查未完成的交易
export function usePendingTransactions() {
  useEffect(() => {
    const keys = Object.keys(localStorage).filter(k => k.startsWith('pending_tx_'));
    for (const key of keys) {
      const orderId = key.replace('pending_tx_', '');
      const txHash = localStorage.getItem(key);
      // 尝试补提交
      api.post(`/v1/orders/${orderId}/confirm`, { txHash })
        .then(() => localStorage.removeItem(key))
        .catch(() => {}); // 后端兜底任务会处理
    }
  }, []);
}
```

## 兜底机制

### 场景覆盖

| 场景 | 问题 | 兜底方案 |
|-----|------|---------|
| 用户关闭页面 | TxHash 未提交 | 定时扫链匹配 |
| 前端提交失败 | 网络问题 | localStorage 重试 + 定时扫链 |
| 订单过期 | 超时未支付 | 定时标记过期 |
| 交易确认慢 | 区块拥堵 | 延迟确认检查 |

### 定时任务设计

```go
// internal/apiserver/job/payment.go

// 每分钟执行：扫描 pending 订单，查链匹配交易
func (j *PaymentJob) ScanPendingOrders(ctx context.Context) error {
    // 1. 查询所有 pending 且未过期的订单
    orders, err := j.orderStore.FindPending(ctx)
    if err != nil {
        return err
    }

    for _, order := range orders {
        // 2. 获取用户钱包地址
        user, _ := j.userStore.Get(ctx, order.UID)
        if user.WalletAddress == "" {
            continue
        }

        // 3. 查询该地址最近的转出交易
        txs, err := j.chainClient.GetTransactions(ctx, order.ChainID, user.WalletAddress, order.CreatedAt)
        if err != nil {
            log.Warnf("scan chain failed: %v", err)
            continue
        }

        // 4. 匹配交易
        for _, tx := range txs {
            if j.matchTransaction(order, tx) {
                j.orderStore.MarkPaid(ctx, order.OrderID, tx.Hash)
                log.Infof("order %s auto confirmed by tx %s", order.OrderID, tx.Hash)
                break
            }
        }
    }

    return nil
}

func (j *PaymentJob) matchTransaction(order *model.Order, tx *Transaction) bool {
    // 收款地址匹配
    if !strings.EqualFold(tx.To, j.cfg.PaymentAddresses[order.ChainID]) {
        return false
    }

    // 金额匹配（允许 1% 误差）
    minAmount := order.Amount.Mul(decimal.NewFromFloat(0.99))
    if tx.Value.LessThan(minAmount) {
        return false
    }

    // 时间匹配（交易时间在订单创建之后）
    if tx.Timestamp.Before(order.CreatedAt) {
        return false
    }

    // 确认数足够
    if tx.Confirmations < j.cfg.RequiredConfirmations[order.ChainID] {
        return false
    }

    // txHash 未被使用
    if j.orderStore.TxHashExists(context.Background(), tx.Hash) {
        return false
    }

    return true
}
```

### 过期订单处理

```go
// 每分钟执行：标记过期订单
func (j *PaymentJob) ExpireOrders(ctx context.Context) error {
    return j.orderStore.MarkExpired(ctx, time.Now())
}

// store 实现
func (s *orderStore) MarkExpired(ctx context.Context, now time.Time) error {
    return s.db.WithContext(ctx).
        Model(&model.Order{}).
        Where("status = ?", model.OrderStatusPending).
        Where("expires_at < ?", now).
        Update("status", model.OrderStatusExpired).
        Error
}
```

### 任务注册

```go
// internal/apiserver/job/register.go

func RegisterJobs(scheduler *asynq.Scheduler, cfg *config.Config) {
    paymentJob := NewPaymentJob(cfg, ...)

    // 每分钟扫描 pending 订单
    scheduler.Register("@every 1m", asynq.NewTask("payment:scan", nil))

    // 每分钟标记过期订单
    scheduler.Register("@every 1m", asynq.NewTask("payment:expire", nil))
}
```

### 链上查询服务

```go
// internal/apiserver/service/chain/client.go

type ChainClient interface {
    GetTransaction(ctx context.Context, chainID int, txHash string) (*Transaction, error)
    GetTransactions(ctx context.Context, chainID int, address string, since time.Time) ([]*Transaction, error)
}
```

### RPC / API 选择

| 方案 | 优点 | 缺点 | 推荐场景 |
|-----|------|------|---------|
| Etherscan API | 简单，有历史交易索引 | 有频率限制（5次/秒） | 订单量小 |
| Alchemy/Infura | 稳定，有 Webhook | 需付费 | 生产环境 |
| 自建节点 | 无限制 | 运维成本高 | 大规模 |

建议：初期用 Etherscan API（免费），量大后切 Alchemy。

## 安全措施

| 层级 | 措施 | 防护 |
|-----|------|------|
| 传输层 | HTTPS | 防中间人 |
| 前端构建 | SRI (Subresource Integrity) | 防 CDN 篡改 |
| 依赖管理 | lockfile + 依赖审计 | 防供应链攻击 |
| 代码层 | 前端硬编码地址 | 增加篡改成本 |
| 用户层 | 提示用户核对钱包显示的地址 | 最终防线 |

## 文件改动清单

| 文件 | 改动类型 | 说明 |
|-----|---------|------|
| `configs/bingo-apiserver.example.yaml` | 修改 | 新增 `payment` 配置块 |
| `pkg/api/apiserver/v1/order.go` | 新增 | 订单相关 Request/Response |
| `internal/apiserver/model/order.go` | 新增 | Order Model |
| `internal/apiserver/model/product.go` | 新增 | Product Model |
| `internal/apiserver/store/order.go` | 新增 | Order Store |
| `internal/apiserver/biz/order/order.go` | 新增 | 订单业务逻辑 |
| `internal/apiserver/handler/order/` | 新增 | 订单 Handler |
| `internal/apiserver/router/order.go` | 新增 | 路由注册 |
| `internal/apiserver/service/chain/` | 新增 | 链上查询服务 |
| `internal/apiserver/job/payment.go` | 新增 | 支付兜底任务 |
| `internal/pkg/errno/payment.go` | 新增 | 支付相关错误码 |
| 数据库迁移 | 新增 | uc_order, uc_product 表 |
