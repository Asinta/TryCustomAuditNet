using System;
using Newtonsoft.Json;

namespace TryCustomAuditNet
{
    public class OrderBase
    {
        public string Name { get; set; }
    }
    
    public class Order : OrderBase
    {
        public Guid Id { get; set; }
        
        public string CustomerName { get; set; }
        
        public int TotalAmount { get; set; }
        public DateTime OrderTime { get; set; }

        public Product Product { get; set; }

        public Order(string name, Guid id, string customerName, int totalAmount, DateTime orderTime, Product product)
        {
            Id = id;
            CustomerName = customerName;
            TotalAmount = totalAmount;
            OrderTime = orderTime;
            Product = product;
            Name = name;
        }

        public void UpdateOrderAmount(int newOrderAmount)
        {
            TotalAmount = newOrderAmount;
        }

        public void UpdateName(string name)
        {
            CustomerName = name;
        }
    }

    public class Product
    {
        [UnAuditable]
        public string ProductName { get; set; }
        
        public int ProductPrice { get; set; }

        public Guid ProductId { get; set; }
    }
}