CREATE TABLE `warehouse_bases` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '倉庫ID',
  `code` varchar(16) NOT NULL COMMENT '倉庫コード',
  `name` varchar(36) NOT NULL COMMENT '倉庫名',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_warehouse_bases_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='倉庫マスタ'
