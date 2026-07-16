# Database Design

## users
id
role
company_name
email
password_hash
phone
city

## products
id
producer_id
name
category
stock
capacity
harvest_date

## rfqs
id
buyer_id
product
quantity
delivery_date
city
status

## demand_groups
id
product
city
total_quantity
delivery_window

## demand_group_members
group_id
rfq_id

## offers
id
group_id
producer_id
price
estimated_delivery

## orders
id
offer_id
status

## reference_prices
commodity
city
price
updated_at

Relations
User(Producer)->Products
User(Buyer)->RFQs
DemandGroup->Many RFQs
DemandGroup->Many Offers
Offer->Order
