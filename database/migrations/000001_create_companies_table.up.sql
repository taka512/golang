CREATE TABLE `companies` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '会社ID',
  `code` varchar(16) NOT NULL COMMENT '会社コード',
  `name` varchar(36) NOT NULL COMMENT '会社名',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_companies_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会社マスタ'
