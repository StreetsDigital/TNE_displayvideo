#!/usr/bin/env python3
from openpyxl import load_workbook
import json

def convert_excel_to_json(excel_file):
    wb = load_workbook(excel_file, data_only=True)
    sheet = wb['PLACEMENTS']

    mapping = {
        "publisher": {
            "publisherId": "icisic-media",
            "domain": "totalprosports.com",
            "defaultBidders": ["rubicon", "kargo", "sovrn", "oms", "aniview", "pubmatic", "triplelift"]
        },
        "adUnits": {}
    }

    # Bidder column mappings
    bidder_columns = {
        "rubicon": {
            "accountId": 6,  # Column F
            "siteId": 7,     # Column G
            "zoneId": 8,     # Column H
            "bidonmultiformat": 9
        },
        "kargo": {
            "placementId": 10  # Column J
        },
        "sovrn": {
            "tagid": 11  # Column K
        },
        "oms": {
            "publisherId": 12  # Column L
        },
        "aniview": {
            "publisherId": 14,  # Column N
            "channelId": 15     # Column O
        },
        "pubmatic": {
            "publisherId": 16,  # Column P
            "adSlot": 17        # Column Q
        },
        "triplelift": {
            "inventoryCode": 18  # Column R
        }
    }

    # Process desktop (rows 3-13) and mobile (rows 19-29)
    for row_idx in list(range(3, 14)) + list(range(19, 30)):
        row = sheet[row_idx]
        slot_pattern = row[0].value  # Column A

        if not slot_pattern or slot_pattern == "Slot Pattern":
            continue

        # Skip if already processed
        if slot_pattern in mapping["adUnits"]:
            continue

        ad_unit_config = {}

        # Extract all bidder parameters
        for bidder, columns in bidder_columns.items():
            bidder_params = {}
            has_data = False

            for param_name, col_idx in columns.items():
                value = row[col_idx - 1].value
                if value is not None and str(value).strip() != "":
                    # Convert float to int for numeric IDs
                    if isinstance(value, float) and param_name != "bidonmultiformat":
                        value = int(value)
                    elif param_name == "bidonmultiformat":
                        value = bool(value)
                    bidder_params[param_name] = value
                    has_data = True

            if has_data:
                ad_unit_config[bidder] = bidder_params

        if ad_unit_config:
            mapping["adUnits"][slot_pattern] = ad_unit_config

    return mapping

if __name__ == '__main__':
    excel_file = 'docs/integrations/tps-onboarding.xlsx'
    mapping = convert_excel_to_json(excel_file)

    with open('config/bizbudding-all-bidders-mapping.json', 'w') as f:
        json.dump(mapping, f, indent=2)

    print(f"✓ Generated config/bizbudding-all-bidders-mapping.json")
    print(f"✓ {len(mapping['adUnits'])} ad units configured")

    # Count bidders
    all_bidders = set()
    for ad_unit in mapping['adUnits'].values():
        all_bidders.update(ad_unit.keys())

    print(f"✓ {len(all_bidders)} bidders: {', '.join(sorted(all_bidders))}")

    print(f"\nSample mapping (first ad unit):")
    first_slot = list(mapping['adUnits'].keys())[0]
    print(f"  {first_slot}:")
    for bidder, params in mapping['adUnits'][first_slot].items():
        print(f"    {bidder}: {params}")
