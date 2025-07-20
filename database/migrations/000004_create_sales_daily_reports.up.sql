 CREATE TABLE `sales_daily_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `company_id` int unsigned NOT NULL COMMENT '会社ID',
  `warehouse_base_id` int unsigned NOT NULL COMMENT '倉庫ID',
  `target_date` date NOT NULL COMMENT '対象日',
  `sales_account_title_id` int unsigned NOT NULL COMMENT '売上科目ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_sales_daily_reports` (`company_id`,`warehouse_base_id`,`target_date`, `sales_account_title_id`),
  KEY `idx_sales_daily_reports_company` (`company_id`,`target_date`,`sales_account_title_id`),
  KEY `idx_sales_daily_reports_warehouse_base` (`warehouse_base_id`,`target_date`,`sales_account_title_id`),
  KEY `idx_sales_daily_reports_date` (`target_date`,`sales_account_title_id`),
  KEY `idx_sales_daily_reports_sales_account_title` (`sales_account_title_id`),
  CONSTRAINT `foreign_sales_daily_reports_company` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`),
  CONSTRAINT `foreign_sales_daily_reports_warehouse_base` FOREIGN KEY (`warehouse_base_id`) REFERENCES `warehouse_bases` (`id`),
  CONSTRAINT `foreign_sales_daily_reports_sales_account_title` FOREIGN KEY (`sales_account_title_id`) REFERENCES `sales_account_titles` (`id`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='日次売上レポート'
