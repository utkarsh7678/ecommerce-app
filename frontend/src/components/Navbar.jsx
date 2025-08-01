import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  AppBar, 
  Toolbar, 
  Typography, 
  Button, 
  IconButton, 
  Badge, 
  Menu, 
  MenuItem,
  Box,
  Avatar,
  Divider,
  ListItemIcon,
  ListItemText,
  Tooltip,
  useTheme,
  useMediaQuery
} from '@mui/material';
import { 
  AccountCircle,
  ExitToApp,
  Menu as MenuIcon
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { toast } from 'react-toastify';

function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [anchorEl, setAnchorEl] = useState(null);
  const [cartAnchorEl, setCartAnchorEl] = useState(null);
  const [cartItems, setCartItems] = useState([]);
  const [loading, setLoading] = useState(false);

  const handleMenu = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleCartOpen = (event) => {
    setCartAnchorEl(event.currentTarget);
    fetchCart();
  };

  const handleCartClose = () => {
    setCartAnchorEl(null);
  };

  const fetchCart = async () => {
    try {
      setLoading(true);
      console.log('Fetching cart data...');
      
      // Use environment variable for API URL with fallback to localhost
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:5000';
      const response = await fetch(`${apiUrl}/api/cart`, {
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('Cart API Error Response:', errorText);
        let errorMessage = `HTTP error! status: ${response.status}`;
        
        try {
          const errorData = JSON.parse(errorText);
          errorMessage = errorData.message || errorMessage;
        } catch (e) {
          console.error('Failed to parse error response:', e);
        }
        
        throw new Error(errorMessage);
      }
      
      const data = await response.json();
      console.log('Raw cart data:', data);
      
      // Handle different response formats
      let items = [];
      if (Array.isArray(data)) {
        items = data; // If the response is directly an array
      } else if (data && Array.isArray(data.items)) {
        items = data.items; // If the response has an items array
      } else if (data && data.products) {
        // If the response has a products array (some APIs use this format)
        items = data.products.map(p => ({
          id: p.id || p._id,
          name: p.name || p.productName,
          price: p.price || 0,
          quantity: p.quantity || 1,
          image: p.image || p.imageUrl || ''
        }));
      }
      
      console.log('Processed cart items:', items);
      setCartItems(items);
      
    } catch (error) {
      console.error('Failed to fetch cart:', error);
      toast.error('Failed to load cart: ' + (error.message || 'Unknown error'));
      setCartItems([]);
    } finally {
      setLoading(false);
    }
  };

  const handleViewOrders = () => {
    handleClose();
    navigate('/orders');
  };

  const handleLogout = () => {
    handleClose();
    logout();
    navigate('/login');
    toast.success('Logged out successfully');
  };

  const cartItemCount = cartItems.reduce((total, item) => total + (item.quantity || 1), 0);
  const cartTotal = cartItems.reduce(
    (total, item) => total + ((item.price || 0) * (item.quantity || 1)), 
    0
  ).toFixed(2);

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const location = useLocation();

  return (
    <AppBar 
      position="fixed" 
      elevation={1}
      sx={{ 
        zIndex: theme.zIndex.drawer + 1,
        backgroundColor: 'primary.main',
        color: 'white',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
        borderBottom: `1px solid ${theme.palette.primary.dark}`
      }}
    >
      <Toolbar disableGutters sx={{ px: { xs: 2, md: 3 }, height: 64 }}>
        <Box 
          sx={{ 
            display: 'flex', 
            alignItems: 'center',
            flexGrow: 1,
            '&:hover': {
              cursor: 'pointer',
              '& .logo-text': {
                transform: 'translateX(2px)'
              }
            }
          }}
          onClick={() => navigate('/')}
        >
          <Typography
            variant="h5"
            className="logo-text"
            sx={{
              fontWeight: 700,
              color: 'white',
              textShadow: '0 1px 2px rgba(0,0,0,0.2)',
              transition: 'transform 0.3s ease',
              letterSpacing: '0.5px',
              display: { xs: 'none', sm: 'block' }
            }}
          >
            E-Commerce Pro
          </Typography>
        </Box>
        
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {/* Profile Menu */}
          <Tooltip title="Account settings">
            <IconButton
              size="medium"
              onClick={handleMenu}
              aria-controls={Boolean(anchorEl) ? 'account-menu' : undefined}
              aria-haspopup="true"
              aria-expanded={Boolean(anchorEl) ? 'true' : undefined}
              sx={{
                p: 0.5,
                ml: 1,
                border: '1px solid',
                borderColor: 'divider',
                color: 'white',
                '&:hover': {
                  backgroundColor: 'rgba(255, 255, 255, 0.1)',
                  color: 'white'
                },
                transition: 'all 0.2s ease-in-out'
              }}
            >
              {user?.avatar ? (
                <Avatar 
                  alt={user.name || 'User'} 
                  src={user.avatar} 
                  sx={{ 
                    width: 36, 
                    height: 36,
                    border: '2px solid',
                    borderColor: 'background.paper',
                    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
                  }}
                />
              ) : (
                <Avatar 
                  sx={{ 
                    bgcolor: 'primary.main',
                    color: 'primary.contrastText',
                    width: 36,
                    height: 36,
                    fontSize: '1.1rem',
                    fontWeight: 600
                  }}
                >
                  {(user?.name || 'U').charAt(0).toUpperCase()}
                </Avatar>
              )}
            </IconButton>
          </Tooltip>
        </Box>
      </Toolbar>

      {/* Cart Menu */}
      <Menu
        anchorEl={cartAnchorEl}
        open={Boolean(cartAnchorEl)}
        onClose={handleCartClose}
        PaperProps={{
          sx: {
            width: 320,
            maxWidth: '100%',
            p: 1,
            mt: 1.5,
            border: '1px solid',
            borderColor: 'primary.main',
            '&:before': {
              content: '""',
              display: 'block',
              position: 'absolute',
              top: 0,
              right: 14,
              width: 10,
              height: 10,
              bgcolor: 'background.paper',
              borderTop: '1px solid',
              borderLeft: '1px solid',
              borderColor: 'primary.main',
              transform: 'translateY(-50%) rotate(45deg)',
              zIndex: 0,
            },
          },
        }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <Box sx={{ 
          p: 2, 
          borderBottom: '1px solid', 
          borderColor: 'primary.main',
          backgroundColor: 'primary.main',
          color: 'white'
        }}>
          <Typography variant="h6" fontWeight={600} color="white">
            Shopping Cart
          </Typography>
          <Typography variant="body2" sx={{ color: 'rgba(35, 156, 43, 0.5)' }}>
            {cartItemCount} {cartItemCount === 1 ? 'item' : 'items'}
          </Typography>
        </Box>
        
        <Box sx={{ maxHeight: 400, overflowY: 'auto', py: 1 }}>
          {loading ? (
            <Box sx={{ p: 2, textAlign: 'center' }}>
              <CircularProgress size={24} />
              <Typography variant="body2" sx={{ mt: 1 }}>Loading cart...</Typography>
            </Box>
          ) : cartItems.length === 0 ? (
            <Box sx={{ p: 3, textAlign: 'center' }}>
              <Box sx={{ 
                width: 64, 
                height: 64, 
                borderRadius: '50%', 
                backgroundColor: 'action.hover',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 16px',
                opacity: 0.5
              }}>
                <Box sx={{ fontSize: 32 }}>🛒</Box>
              </Box>
              <Typography variant="body1" color="text.secondary">
                Your cart is empty
              </Typography>
              <Button 
                variant="contained" 
                color="primary" 
                size="small" 
                sx={{ mt: 2 }}
                onClick={() => {
                  handleCartClose();
                  navigate('/');
                }}
              >
                Continue Shopping
              </Button>
            </Box>
          ) : (
            <>
              {cartItems.map((item) => {
                // Debug log for each item
                console.log('Rendering cart item:', item);
                
                // Safely access item properties with defaults
                const itemName = item.name || item.productName || 'Unnamed Product';
                const itemPrice = Number(item.price) || 0;
                const itemQuantity = Number(item.quantity) || 1;
                const itemImage = item.image || '';
                
                return (
                  <MenuItem 
                    key={item.id || Math.random().toString(36).substr(2, 9)}
                    sx={{ 
                      borderRadius: 1,
                      mb: 0.5,
                      color: 'black',
                      '&:hover': {
                        backgroundColor: 'rgba(0, 0, 0, 0.04)',
                      }
                    }}
                  >
                    <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                      <Avatar 
                        src={itemImage} 
                        alt={itemName}
                        variant="rounded"
                        sx={{ 
                          width: 48, 
                          height: 48, 
                          mr: 2,
                          backgroundColor: 'rgba(0, 0, 0, 0.1)'
                        }}
                      >
                        {!itemImage && itemName.charAt(0).toUpperCase()}
                      </Avatar>
                      <Box sx={{ flexGrow: 1, minWidth: 0, overflow: 'hidden' }}>
                        <Typography 
                          variant="subtitle2" 
                          noWrap 
                          sx={{ 
                            fontWeight: 600, 
                            mb: 0.5,
                            color: 'black',
                            fontSize: '0.875rem',
                            lineHeight: 1.2
                          }}
                        >
                          {itemName}
                        </Typography>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Typography variant="body2" sx={{ color: 'text.secondary', fontSize: '0.8rem' }}>
                            {itemQuantity} × ${itemPrice.toFixed(2)}
                          </Typography>
                          <Typography variant="subtitle2" fontWeight={600} sx={{ fontSize: '0.9rem' }}>
                            ${(itemPrice * itemQuantity).toFixed(2)}
                          </Typography>
                        </Box>
                      </Box>
                    </Box>
                  </MenuItem>
                );
              })}
              
              <Box sx={{ p: 2, borderTop: '1px solid', borderColor: 'divider', mt: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                  <Typography variant="subtitle1" fontWeight={600}>
                    Subtotal:
                  </Typography>
                  <Typography variant="subtitle1" fontWeight={600}>
                    ${cartTotal}
                  </Typography>
                </Box>
                <Button
                  fullWidth
                  variant="contained"
                  color="primary"
                  onClick={() => {
                    handleCartClose();
                    navigate('/checkout');
                  }}
                  sx={{
                    py: 1,
                    textTransform: 'none',
                    fontWeight: 600,
                    borderRadius: 2,
                    boxShadow: 'none',
                    '&:hover': {
                      boxShadow: '0 4px 12px rgba(25, 118, 210, 0.2)',
                      transform: 'translateY(-1px)',
                    },
                    transition: 'all 0.2s ease-in-out',
                  }}
                >
                  Checkout
                </Button>
                <Button
                  fullWidth
                  variant="outlined"
                  color="primary"
                  onClick={() => {
                    handleCartClose();
                    navigate('/cart');
                  }}
                  sx={{
                    mt: 1,
                    py: 1,
                    textTransform: 'none',
                    borderRadius: 2,
                  }}
                >
                  View Cart
                </Button>
              </Box>
            </>
          )}
        </Box>
      </Menu>

      {/* User Menu */}
      <Menu
        anchorEl={anchorEl}
        id="account-menu"
        open={Boolean(anchorEl)}
        onClose={handleClose}
        onClick={handleClose}
        PaperProps={{
          elevation: 3,
          sx: {
            overflow: 'visible',
            filter: 'drop-shadow(0px 4px 12px rgba(0,0,0,0.1))',
            mt: 1.5,
            minWidth: 220,
            '& .MuiAvatar-root': {
              width: 32,
              height: 32,
              ml: -0.5,
              mr: 1,
            },
            '&:before': {
              content: '""',
              display: 'block',
              position: 'absolute',
              top: 0,
              right: 14,
              width: 10,
              height: 10,
              bgcolor: 'background.paper',
              transform: 'translateY(-50%) rotate(45deg)',
              zIndex: 0,
            },
          },
        }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <Box sx={{ px: 2, py: 1.5, borderBottom: '1px solid', borderColor: 'divider', minWidth: '200px' }}>
          <Typography variant="subtitle1" fontWeight={600} color="text.primary">
            {user?.name || 'User'}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, fontSize: '0.8rem' }}>
            {user?.email || ''}
          </Typography>
        </Box>
        <Divider />
        <MenuItem onClick={handleLogout}>
          <ListItemIcon>
            <ExitToApp fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>Logout</ListItemText>
        </MenuItem>
      </Menu>
    </AppBar>
  );
}

export default Navbar;