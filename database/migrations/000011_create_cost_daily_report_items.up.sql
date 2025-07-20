CREATE TABLE `cost_daily_report_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cost_daily_report_id` bigint unsigned NOT NULL COMMENT '日次原価レポートID',
  `size` varchar(16) DEFAULT NULL COMMENT 'サイズ',
  `quantity` int NOT NULL DEFAULT '0' COMMENT '数量',
  `cost_price` decimal(9,3) NOT NULL DEFAULT '0.000' COMMENT '単価',
  `cost_amount` decimal(13,3) NOT NULL DEFAULT '0.000' COMMENT '金額',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_cost_daily_report_items_id` (`cost_daily_report_id`),
  CONSTRAINT `foreign_cost_daily_report_items_id` FOREIGN KEY (`cost_daily_report_id`) REFERENCES `cost_daily_reports` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日次原価レポート明細'
