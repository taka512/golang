CREATE TABLE `sales_daily_report_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `sales_daily_report_id` bigint unsigned NOT NULL COMMENT '日次売上レポートID',
  `size` varchar(16) DEFAULT NULL COMMENT 'サイズ',
  `quantity` int NOT NULL DEFAULT '0' COMMENT '数量',
  `price` decimal(9,3) NOT NULL DEFAULT '0.000' COMMENT '単価',
  `amount` decimal(13,3) NOT NULL DEFAULT '0.000' COMMENT '請求金額',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sales_daily_report_items_id` (`sales_daily_report_id`),
  CONSTRAINT `foreign_sales_daily_report_items_id` FOREIGN KEY (`sales_daily_report_id`) REFERENCES `sales_daily_reports` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COMMENT='日次売上レポート明細'
