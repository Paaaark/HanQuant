import pandas as pd

# Fixed-width fields and sizes matching your Go parser
field_widths = [9, 12, 40, 2, 1, 4, 4, 4]

field_names = [
    "Code",
    "ISIN",
    "Name",
    "SecurityType",
    "CapSize",
    "IndLarge",
    "IndMedium",
    "IndSmall"
]

def parse_mst(filepath):
    part1_rows = []
    part2_rows = []

    with open(filepath, mode='r', encoding='cp949') as f:
        for row in f:
            part1_str = row[0:len(row) - 228]
            code = part1_str[0:9].rstrip()
            isin = part1_str[9:21].rstrip()
            name = part1_str[21:].strip()
            part1_rows.append([code, isin, name])

            part2_rows.append(row[-228:].rstrip('\n'))

    # DataFrame for part1
    df_part1 = pd.DataFrame(part1_rows, columns=['Code', 'ISIN', 'Name'])

    # Define fixed-width specs for part2 (same as before)
    field_widths = [2,1,4,4,4,
                    1,1,1,1,
                    1,1,1,1,
                    1,1,1,1,
                    1,1,1,1,1,1,
                    1,1,1,1,
                    1,1,1,1,
                    9,5,5,
                    1,1,1,2,
                    1,1,1,2,
                    2,2,3,1,
                    3,12,12,8,
                    15,21,2,7,
                    1,1,1,1,1,
                    9,9,9,5,9,
                    8,9,3,1,1,1]

    part2_columns = ['GroupCode', 'CapSize', 'IndLarge', 'IndMedium', 'IndSmall', # 2 1 4 4 4
                     'MfgYn', 'LowLiquidity', 'GovernanceYn', 'KOSPI200Sector',  # 1 1 1 1
                     'KOSPI100Yn', 'KOSPI50Yn', 'KRXYn', 'ETPType', # 1 1 1 1
                     'ELWYn', 'KRX100Yn', 'KRXCarYn', 'KRXSemiconductorYn', # 1 1 1 1
                     'KRXBioYn', 'KRXBankYn', 'SPACYn', 'KRXEnergyYn', 'KRXSteelYn', 'ShortOverheatCode', # 1 1 1 1 1 1
                     'KRXMediaYn', 'KRXConstructionYn', 'Non1', 'KRXSecurityYn', # 1 1 1 1
                     'KRXShipYn', 'KRXInsuranceYn', 'KRXTransportYn', 'SRIYn', # 1 1 1 1
                     'BasePrice', 'RegMarketQty', 'AfterHoursQty', # 9 5 5
                     'TradeStopYn', 'SettlementYn', 'ManagedYn', 'MarketAlertCode', # 1 1 1 2
                     'AlertRiskYn', 'NonDisclosureYn', 'BypassListingYn', 'LockCode', # 1 1 1 2
                     'ParValueChangeCode', 'CapitalIncreaseCode', 'MarginRate', 'CreditAvailable', # 2 2 3 1 
                     'CreditPeriod', 'PrevTradeVol', 'FaceValue', 'ListingDate', # 3 12 12 8
                     'ListedShares', 'Capital', 'ClosingMonth', 'IPOPrice', # 15 21 2 7
                     'PreferredStockCode', 'ShortSellOverheatYn', 'AbnormalRiseYn', 'KRX300Yn', 'KOSPIYn', # 1 1 1 1 1
                     'Sales', 'OperatingProfit', 'OrdinaryProfit', 'NetProfit', 'ROE', # 9 9 9 5 9
                     'BaseDate', 'MarketCap', 'GroupCompanyCode', 'CreditLimitOverYn', 'CollateralLoanYn', 'MarginLoanYn'] # 8 9 3 1 1 1
    
    print(len(field_widths), len(part2_columns))


    # Parse fixed-width part2 into DataFrame
    df_part2 = pd.read_fwf(pd.io.common.StringIO('\n'.join(part2_rows)), widths=field_widths, names=part2_columns)

    # Combine by index
    df = pd.concat([df_part1, df_part2], axis=1)

    selected_cols = ["Code", "ISIN", "Name", "GroupCode", "CapSize", "IndLarge", "IndMedium", "IndSmall", "MarketCap"]  # example subset
    return df[selected_cols]
if __name__ == "__main__":
    df1 = parse_mst(".kis_data/kospi_code.mst")
    df2 = parse_mst(".kis_data/kosdaq_code.mst")
    df = pd.concat([df1, df2], ignore_index=True)
    df.to_csv(".kis_data/stock_listings_complete.csv", index=False, encoding="utf-8-sig")
    print("Generated kospi_code_minimal.csv")
