using System;

namespace TryCustomAuditNet
{
    [AttributeUsage(AttributeTargets.Property)]
    public class UnAuditableAttribute: Attribute
    {
    }
}