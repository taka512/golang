# Sale Cost Profit Report

A Go command-line tool that generates comprehensive sale cost profit reports by analyzing data from the sales and cost daily reports tables.

## Overview

This tool queries the database to generate profit reports that combine sales and cost data, calculating profit amounts and margins for companies and warehouse bases over specified date ranges.

## Database Schema

The tool works with the following database tables:

### Sales Tables
- `sales_daily_reports` - Daily sales report headers
- `sales_daily_report_items` - Detailed sales items with amounts
- `sales_account_titles` - Sales account classifications

### Cost Tables  
- `cost_daily_reports` - Daily cost report headers
- `cost_daily_report_items` - Detailed cost items with amounts
- `cost_account_titles` - Cost account classifications

### Reference Tables
- `companies` - Company information
- `warehouse_bases` - Warehouse/location information

## Features

- **Date Range Filtering**: Generate reports for specific date ranges
- **Company & Warehouse Breakdown**: Detailed analysis by company and warehouse
- **Profit Calculations**: Automatic calculation of profit amounts and margins
- **CSV Export**: Export results to CSV format for further analysis
- **Console Summary**: Display summary statistics in the terminal
- **Flexible Date Input**: Default to current month or specify custom ranges

## Usage

### Basic Usage (Current Month)
```bash
make run
```

### Custom Date Range
```bash
make run-range START=2024-01-01 END=2024-01-31
```

### Direct Binary Usage
```bash
# Build first
make build

# Run with default date range
./bin/sale-cost-profit-report

# Run with custom date range
./bin/sale-cost-profit-report 2024-01-01 2024-01-31
```

## Configuration

### Database Connection
Update the DSN (Data Source Name) in `main.go`:
```go
dsn := "username:password@tcp(host:port)/database_name?parseTime=true"
```

### Environment Variables
You can also set database connection via environment variables:
- `DB_HOST` - Database host
- `DB_PORT` - Database port  
- `DB_USER` - Database username
- `DB_PASS` - Database password
- `DB_NAME` - Database name

## Output

### CSV Report
The tool generates a CSV file with the following columns:
- Company ID & Name
- Warehouse ID & Name  
- Target Date
- Sales Amount
- Cost Amount
- Profit Amount
- Profit Margin (%)

### Console Summary
Displays:
- Total records processed
- Overall sales, costs, and profit totals
- Overall profit margin
- Breakdown by company

## Building

### Local Build
```bash
make build
```

### Docker Build
```bash
make docker-build
make docker-run
```

## Dependencies

- Go 1.21+
- MySQL driver: `github.com/go-sql-driver/mysql`

Install dependencies:
```bash
make deps
```

## Examples

### Sample Output
```
Generating Sale Cost Profit Report from 2024-01-01 to 2024-01-31

=== PROFIT REPORT SUMMARY ===
Total Records: 45
Total Sales: 125000.000
Total Costs: 87500.000
Total Profit: 37500.000
Overall Profit Margin: 30.00%

=== BY COMPANY ===
Company A: Sales=75000.000, Costs=52500.000, Profit=22500.000 (30.00%)
Company B: Sales=50000.000, Costs=35000.000, Profit=15000.000 (30.00%)

Report saved to: sale_cost_profit_report_2024-01-01_to_2024-01-31.csv
```

## Troubleshooting

### Database Connection Issues
1. Verify database credentials and connection string
2. Ensure MySQL server is running and accessible
3. Check firewall settings and network connectivity

### No Data Found
1. Verify date range contains data in the sales/cost tables
2. Check that companies and warehouse_bases tables have data
3. Ensure foreign key relationships are properly established

### Performance Optimization
For large datasets:
1. Add appropriate indexes on date columns
2. Consider partitioning large tables by date
3. Use date range limits to avoid processing too much data at once

## Development

### Adding New Features
1. Modify the SQL query in `generateProfitReport()` function
2. Update the `SaleCostProfitReport` struct if adding new fields
3. Update CSV headers and row generation accordingly
4. Add tests for new functionality

### Testing
```bash
make test
```

## License

This tool is part of the larger Go project and follows the same licensing terms.
