CREATE TABLE `company_account_titles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(16) NOT NULL COMMENT '科目コード',
  `name` varchar(32) NOT NULL COMMENT '科目名',
  `disabled` tinyint(4) NOT NULL DEFAULT '0' COMMENT '無効フラグ 0:有効 1:無効',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_company_account_titles_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='請求科目'
