CREATE TABLE `test_reports` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `target_date` date NOT NULL,
  `user_id` int NOT NULL,
  `quantity` int NOT NULL DEFAULT '0',
  `amount` int NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  KEY `idx_tr_date` (`target_date`),
  KEY `idx_tr_id_date` (`user_id`, `target_date`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

