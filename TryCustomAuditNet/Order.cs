using System;

namespace TryCustomAuditNet
{
    public class Order
    {
        public Guid Id { get; set; }
        
        [UnAuditable]
        public string CustomerName { get; set; }
        
        public int TotalAmount { get; set; }
        public DateTime OrderTime { get; set; }

        public Order(Guid id, string customerName, int totalAmount, DateTime orderTime)
        {
            Id = id;
            CustomerName = customerName;
            TotalAmount = totalAmount;
            OrderTime = orderTime;
        }

        public void UpdateOrderAmount(int newOrderAmount)
        {
            TotalAmount = newOrderAmount;
        }
    }
}